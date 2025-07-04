package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
)

type bossPage struct {
	windowPage
}

func newBossPage() bossPage {
	return bossPage{}
}

func (m bossPage) Init() tea.Cmd {
	return messages.Post(messages.FooterStatusMessage{HiddenFooter: true})
}

func (m bossPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.windowPage = m.windowPage.Update(msg)
	return m, nil
}

func (m bossPage) View() string {
	// 深色背景 + 居中提示文字
	shieldStyle := lipgloss.NewStyle().
		Width(m.w).
		Height(m.h).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)
	return shieldStyle.Render("")
}
