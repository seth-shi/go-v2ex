package messages

import (
	tea "github.com/charmbracelet/bubbletea"
)

func ErrorOrToast(fn func() error, text string) tea.Cmd {

	return func() tea.Msg {

		if err := fn(); err != nil {
			return err
		}

		if text == "" {
			return nil
		}

		return ShowToastRequest{Text: text}
	}
}

func Post(msg tea.Msg) tea.Cmd {

	if msg == nil {
		return nil
	}

	return func() tea.Msg {
		return msg
	}
}
