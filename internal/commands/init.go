package commands

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/pkg"
)

func DispatchConfigLoaded(err error) tea.Cmd {
	return func() tea.Msg {
		slog.Info("配置加载完成", slog.Any("err", err))
		return messages.LoadConfigResult{Error: err}
	}
}
func CheckAppHasNewVersion(currVersion string) tea.Cmd {
	return func() tea.Msg {
		result := pkg.CheckLatestRelease(currVersion)
		if result == nil {
			return nil
		}

		return messages.ShowToastRequest{Text: result.GetTitle()}
	}
}
