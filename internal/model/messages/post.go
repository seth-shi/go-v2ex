package messages

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Post(msg tea.Msg) tea.Cmd {

	if msg == nil {
		return nil
	}

	return func() tea.Msg {
		return msg
	}
}
