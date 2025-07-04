package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/pkg"
)

func CheckAppHasNewVersion(currVersion string) tea.Cmd {
	return func() tea.Msg {
		result := pkg.CheckLatestRelease(currVersion)
		if result == nil {
			return nil
		}

		return messages.ShowToastRequest{Text: result.GetTitle()}
	}
}
