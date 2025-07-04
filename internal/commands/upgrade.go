package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v73/github"
	"github.com/mholt/archives"
	"github.com/minio/selfupdate"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/pkg"
)

// 下载进度
var (
	upgrading    atomic.Bool
	errUpgrading = errors.New("正在升级中")
)

func UpgradeApp(appVersion string) tea.Cmd {
	return func() tea.Msg {

		if !upgrading.CompareAndSwap(false, true) {
			return errUpgrading
		}
		defer upgrading.Store(false)

		result := pkg.CheckLatestRelease(appVersion)
		if result == nil {
			return errors.New("无可更新的应用版本")
		}
		latestResult := result.Result

		slog.Info("github 应用", slog.Any("asserts", latestResult.Assets))
		// name=go-v2ex-${{ matrix.goos }}-${{ matrix.goarch }}
		name := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
		for _, asset := range latestResult.Assets {
			if strings.Contains(asset.GetName(), name) {
				// 发送下载消息, 去更新应用
				state := messages.NewDownloadState(uint64(asset.GetSize()))
				go upgradeByAsset(state, asset)
				return messages.UpgradeStateMessage{State: state}
			}
		}
		// 跳转到新页面
		return fmt.Errorf("无当前系统适配文件,点击[%s]去查找", latestResult.GetAssetsURL())
	}
}

func upgradeByAsset(state *messages.UpgradeState, asset *github.ReleaseAsset) {
	// 移除正常流程中的锁，只在设置错误时加锁
	var (
		err          error
		dstDir       string
		compressFile string
		executeFile  string
		newPf        *os.File
	)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
		state.SetError(err)
	}()

	// 后续的操作都在这个目录操作
	dstDir, err = os.MkdirTemp("", fmt.Sprintf("*-%s", consts.AppName))
	if err != nil {
		return
	}
	defer os.RemoveAll(dstDir)

	// 下载文件
	state.SetState(messages.UpgradeStateDownloading, "开始下载文件")
	compressFile, err = downloadAsset(dstDir, state, asset)
	if err != nil {
		return
	}

	// 开始解压文件夹
	state.SetState(messages.UpgradeStateExtracting, "开始解压文件")
	executeFile, err = unpackFile(dstDir, compressFile)
	if err != nil {
		return
	}

	// 开始替换当前可执行文件
	state.SetState(messages.UpgradeFinalStep, "开始替换可执行文件")
	newPf, err = os.Open(executeFile)
	if err != nil {
		return
	}
	defer newPf.Close()

	err = selfupdate.Apply(newPf, selfupdate.Options{})
	if err != nil {
		return
	}
	for i := 3; i > 0; i-- {
		state.SetState(messages.UpgradeFinalStep, fmt.Sprintf("程序替换完成,%d秒后程序退出", i))
		time.Sleep(time.Second)
	}
	state.SetState(messages.UpgradeStateFinished, "")
}

func unpackFile(dir string, downloadFile string) (string, error) {
	// 获取当前可执行文件的路径
	var (
		ctx = context.Background()
	)
	format, _, err := archives.Identify(ctx, downloadFile, nil)
	if err != nil {
		return "", err
	}

	ex, ok := format.(archives.Extractor)
	if !ok {
		return "", fmt.Errorf("无效的压缩格式%s", format.Extension())
	}

	pf, err := os.Open(downloadFile)
	if err != nil {
		return "", err
	}

	// 只有一个解压文件, 处理第一个就行
	var (
		executeFile string
		isFirstFile bool
	)
	err = ex.Extract(
		ctx, pf, func(ctx context.Context, info archives.FileInfo) error {

			if !isFirstFile {
				isFirstFile = true
			}

			pf, err := info.Open()
			if err != nil {
				return err
			}
			defer pf.Close()

			// 创建目标文件
			executeFile = filepath.Join(dir, info.Name())
			outFile, err := os.Create(executeFile)
			if err != nil {
				return err
			}
			defer outFile.Close()

			// 将文件内容从压缩包复制到目标文件
			_, err = io.Copy(outFile, pf)
			return err
		},
	)

	return executeFile, err
}

func downloadAsset(dir string, state *messages.UpgradeState, asset *github.ReleaseAsset) (string, error) {
	// 下载的文件
	body, _, err := pkg.DownloadAsset(asset)
	if err != nil {
		return "", err
	}
	defer body.Close()

	// 下载到压缩包到临时文件
	tmpPf, err := os.Create(path.Join(dir, asset.GetName()))
	if err != nil {
		return "", err
	}
	defer tmpPf.Close()

	// 复制数据并显示进度
	_, err = io.Copy(tmpPf, io.TeeReader(body, state))
	if err != nil {
		return "", err
	}

	return tmpPf.Name(), nil
}

// 定时查询下载进度
func CheckDownloadProcessMessages(state *messages.UpgradeState) func(t time.Time) tea.Msg {
	return func(t time.Time) tea.Msg {

		// 如果发生错误, 那么就不要发消息了
		// 如果下载完成了
		if err := state.Error(); err != nil {
			// 开始移动文件
			// 解压 zip 包
			return messages.ShowAlertRequest{Text: err.Error()}
		}

		if state.Finished() {
			// 开始移动文件
			// 解压 zip 包
			return tea.Quit()
		}

		// 否则返回下载进度条, 并且下一次继续显示进度条
		return messages.UpgradeStateMessage{State: state}
	}
}
