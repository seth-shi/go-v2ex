package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) headerView() string {

	var (
		leftText = ""
		nickname = ""
	)

	if m.ctx.Me == nil {
		nickname = "未登录"
		if m.ctx.Config.Token != "" {
			nickname = "登录中..."
		}
	} else {
		nickname = m.ctx.Me.Username
	}

	m.ctx.ContentHeight = 0
	padding := 1
	view := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		Width(m.ctx.ScreenWidth).
		PaddingBottom(padding).
		PaddingLeft(padding).
		PaddingRight(padding).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				leftText,
				lipgloss.PlaceHorizontal(
					m.ctx.ScreenWidth-lipgloss.Width(leftText)-2*padding,
					lipgloss.Right,
					nickname,
				),
			),
		)
	m.ctx.ContentHeight += lipgloss.Height(view)
	return view
}
