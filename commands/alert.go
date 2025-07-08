package commands

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"go.dalton.dog/bubbleup"
)

const (
	alertPerSeconds = 2
)

// 内部只能通过实例去发送消息, 所以无法任意地方调用, 这里改一下
// 多个消息延迟发送
var (
	alert       = bubbleup.NewAlertModel(80, true)
	alertLocker sync.RWMutex
	alertLastAt int64
)

func Alert(key, msg string) tea.Cmd {

	// 如果现在时间超过上一次发送时间+2, 那么直接用现在时间
	alertLocker.Lock()
	var (
		now    = time.Now().Unix()
		sendAt = alertLastAt + alertPerSeconds
	)
	if sendAt < now {
		sendAt = now
	}
	alertLastAt = sendAt
	alertLocker.Unlock()

	// 如果
	diffSeconds := sendAt - now

	if diffSeconds <= 0 {
		return alert.NewAlertCmd(key, msg)
	}

	diff := time.Duration(diffSeconds) * time.Second
	return tea.Tick(
		diff, func(t time.Time) tea.Msg {
			return alert.NewAlertCmd(key, msg)()
		},
	)
}

func AlertError(err error) tea.Cmd {
	if err == nil {
		return nil
	}

	return Alert(bubbleup.ErrorKey, err.Error())
}

func AlertInfo(msg string) tea.Cmd {
	return Alert(bubbleup.InfoKey, msg)
}

func AlertWarn(msg string) tea.Cmd {
	return Alert(bubbleup.WarnKey, msg)
}
