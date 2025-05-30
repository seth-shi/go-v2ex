package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	notFoundView = lipgloss.
		NewStyle().
		Height(5).
		Bold(true).
		Padding(2).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#ff5722")).
		SetString("载入中...")
)

func (m Model) bodyView() string {

	view := notFoundView.Width(m.ctx.ScreenWidth).Render()
	if m.currBodyModel != nil {
		view = m.currBodyModel.View()
	}

	m.ctx.ContentHeight += lipgloss.Height(view)
	return view
}

func (m Model) bodyUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {

	// update body message
	if m.currBodyModel != nil {
		var cmd tea.Cmd
		m.currBodyModel, cmd = m.currBodyModel.Update(msg)
		return m, cmd
	}

	return m, nil
}
