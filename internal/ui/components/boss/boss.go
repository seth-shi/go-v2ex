package boss

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
	// 深色背景 + 居中提示文字
	shieldStyle := lipgloss.NewStyle().
		Width(config.Screen.Width).
		Height(config.Screen.Height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)
	return shieldStyle.Render("")
}
