package messages

import (
	"fmt"
	"sync"

	"github.com/seth-shi/go-v2ex/internal/pkg"
)

type UpgradeStateMessage struct {
	State *UpgradeState
}

type UpgradeAppState int

const (
	UpgradeStateInit        UpgradeAppState = iota
	UpgradeStateDownloading UpgradeAppState = iota
	UpgradeStateExtracting
	UpgradeFinalStep
	UpgradeStateFinished
)

type UpgradeState struct {
	downloaded uint64
	total      uint64
	state      UpgradeAppState
	text       string
	err        error
	locker     sync.RWMutex
}

func NewDownloadState(total uint64) *UpgradeState {
	return &UpgradeState{
		total: total,
		state: UpgradeStateInit,
		text:  "初始化更新中",
	}
}

func (wc *UpgradeState) Finished() bool {
	wc.locker.Lock()
	defer wc.locker.Unlock()
	return wc.state == UpgradeStateFinished
}

func (wc *UpgradeState) Error() error {
	wc.locker.Lock()
	defer wc.locker.Unlock()

	return wc.err
}

func (wc *UpgradeState) SetError(err error) {
	wc.locker.Lock()
	defer wc.locker.Unlock()

	wc.err = err
}

func (wc *UpgradeState) SetState(state UpgradeAppState, text string) {
	wc.locker.Lock()
	defer wc.locker.Unlock()

	wc.state = state
	wc.text = text
}

func (wc *UpgradeState) Text() string {
	wc.locker.Lock()
	defer wc.locker.Unlock()

	switch wc.state {
	case UpgradeStateDownloading:
		return fmt.Sprintf("正在下载%.2fMB/%.2fMB", pkg.BytesToMB(int(wc.downloaded)), pkg.BytesToMB(int(wc.total)))
	default:
		return wc.text
	}
}

func (wc *UpgradeState) Write(p []byte) (int, error) {

	wc.locker.Lock()
	defer wc.locker.Unlock()

	n := len(p)
	wc.downloaded += uint64(n)

	return n, nil
}
