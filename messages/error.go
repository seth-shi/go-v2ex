package messages

import (
	tea "github.com/charmbracelet/bubbletea"
)

func ErrorOrToast(err error, msg string) tea.Msg {

	if err != nil {
		return err
	}

	if msg == "" {
		return nil
	}

	return ProxyShowToastRequest{Text: msg}

}
