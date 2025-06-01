package splash

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/config"
)

type Model struct {
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	return m, nil
}

func (m Model) View() string {
	return lipgloss.
		NewStyle().
		Width(config.Screen.Width).
		Bold(true).
		Height(1).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#ff5722")).
		Render("载入中...")
}
