package pkg

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	base8_bagua "github.com/chyroc/base8-bagua"
	"github.com/google/go-github/v73/github"
	"github.com/hashicorp/go-version"
	"github.com/seth-shi/go-v2ex/v2/consts"
)

const (
	GithubOwner    = "seth-shi"
	GithubRepoName = "go-v2ex"
)

var (
	latestAppVersion *github.RepositoryRelease
	versionOnce      sync.Once
	githubClient     = github.
				NewClient(&http.Client{Timeout: time.Second * 5}).
				WithAuthToken(baguaString())
)

type GithubReleaseResult struct {
	Result *github.RepositoryRelease
}

func baguaString() string {
	t := "☳☱☶☶☴☵☶☴☳☲☰☷☲☵☴☲☲☷☶☷☰☱☴☱☳☵☰☵☷☴☶☱☱☴☲☴☰☵☰☷☲☵☴☳☲☱☳☰☲☰☲☵☴☴☶☰☳☶☴☶☰☵☶☱☳☴☴☵☱☱☷☰☳☱☲☷☴☱☴☴☲☲☴☷☴☵☱☵☲☷☶☵☳☵☲☲☳☰☶☴☶☴☶☰☲☳☰☷☲☱☲☰☳☳☴☷☵☱☱☱☳☶☲☵☱☵☱☰☳☳☰☵☳☵☷☱☳☶☴☶☳☵☲☴☳☱☲☳☳☵☳☱☲☱☲☳☰☱☷☰☲☴☴☶☲☵☱☲☲☴☰☶☶☱☵☳☳☱☰☷☲☵☵☶☲☱☶☴☶☵☴☲☳☶☰☴☲☵☶☴☱☵☰☴☴☱☳☰☲☶☰☴☱☱☲☱☲☱☶☵☴☴☶☶☲☳☶☳☱☱☵☶☳☳☴☳☲☱☷☰☲☴☲☳☰☱☴☷"
	tt, _ := base8_bagua.Decode(t)
	return string(tt)
}

func NewGithubReleaseResult(result *github.RepositoryRelease) *GithubReleaseResult {
	return &GithubReleaseResult{Result: result}
}

func (r *GithubReleaseResult) GetTitle() string {
	key := strings.Join(consts.AppKeyMap.UpgradeApp.Keys(), " ")
	return fmt.Sprintf(
		"!!!有新版本:[%s] 按[%s]更新 %s",
		r.Result.GetTagName(),
		key,
		r.Result.GetBody(),
	)
}

func DownloadAsset(asset *github.ReleaseAsset) (io.ReadCloser, string, error) {
	return githubClient.Repositories.DownloadReleaseAsset(
		context.Background(),
		GithubOwner,
		GithubRepoName,
		asset.GetID(),
		&http.Client{Timeout: time.Second * 300},
	)
}

func CheckLatestRelease(currentVersion string) *GithubReleaseResult {
	appVersion, err := version.NewVersion(currentVersion)
	if err != nil {
		slog.Info("应用版本解析失败", slog.Any("err", err))
		return nil
	}

	latestResult := getLatestRelease()
	if latestResult == nil {
		return nil
	}

	latestVersion, err := version.NewVersion(latestResult.GetTagName())
	if err != nil {
		slog.Info("Github 应用版本解析失败", slog.Any("err", err))
		return nil
	}

	if !latestVersion.GreaterThan(appVersion) {
		return nil
	}

	// 开始更新
	return NewGithubReleaseResult(latestResult)
}

func getLatestRelease() *github.RepositoryRelease {
	// 获取最新版本
	versionOnce.Do(
		func() {

			result, _, err := githubClient.Repositories.GetLatestRelease(
				context.Background(), GithubOwner, GithubRepoName,
			)
			if err != nil {
				slog.Info("请求 Github 失败", slog.Any("err", err))
				return
			}

			latestAppVersion = result
		},
	)

	return latestAppVersion
}
