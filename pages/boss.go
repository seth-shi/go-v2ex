package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/g"
)

type bossPage struct {
}

func newBossPage() bossPage {
	return bossPage{}
}

func (m bossPage) Init() tea.Cmd {
	return func() tea.Msg {
		g.Session.HideFooter.Store(true)
		return nil
	}
}

func (m bossPage) Close() error {
	g.Session.HideFooter.Store(false)
	return nil
}

func (m bossPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m bossPage) View() string {

	var (
		width, height = g.Window.GetSize()
	)

	// 深色背景 + 居中提示文字
	shieldStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)
	return shieldStyle.Render("")
}
