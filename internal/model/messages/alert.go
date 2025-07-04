package messages

import (
	tea "github.com/charmbracelet/bubbletea"
	"go.dalton.dog/bubbleup"
)

var (
	alert = bubbleup.NewAlertModel(80, true)
)

// Alert 尽量使用 commands.Alert, 因为它可以使用队列延迟发送
// 下面这个会立即发送
func Alert(key, msg string) tea.Msg {
	// 内部只能通过实例去发送消息, 所以无法任意地方调用, 这里改一下
	return alert.NewAlertCmd(key, msg)()
}

func AlertInfo(msg string) tea.Msg {
	// 内部只能通过实例去发送消息, 所以无法任意地方调用, 这里改一下
	return alert.NewAlertCmd(bubbleup.InfoKey, msg)()
}
