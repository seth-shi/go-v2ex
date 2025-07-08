package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/messages"
	"github.com/seth-shi/go-v2ex/pkg"
)

func CheckAppHasNewVersion(currVersion string) tea.Cmd {
	return func() tea.Msg {
		result := pkg.CheckLatestRelease(currVersion)
		if result == nil {
			return nil
		}

		return messages.ProxyShowToastRequest{Text: result.GetTitle()}
	}
}
