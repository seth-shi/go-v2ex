package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/consts"
)

const (
	rightText = "Powered by seth-shi"
)

func (m Model) footerView() string {

	var (
		leftSection  = lipgloss.NewStyle().SetString("")
		rightSection = lipgloss.NewStyle().SetString(rightText)
	)

	if m.ctx.Error != nil {
		leftSection = leftSection.
			Foreground(lipgloss.Color("#ff5722")).
			SetString(m.ctx.Error.Error())
	} else if m.ctx.LoadingText != nil {
		leftSection = leftSection.SetString(
			fmt.Sprintf(
				"%s %s",
				lipgloss.NewStyle().PaddingLeft(1).Render(
					m.spinner.View(),
				),
				lo.FromPtr(m.ctx.LoadingText),
			),
		)
	} else if m.ctx.TopicPage > 0 {
		leftSection = leftSection.SetString(fmt.Sprintf("第%d页", m.ctx.TopicPage))
	} else {
		helpKey := consts.AppKeyMap.Help.Help()
		leftSection = leftSection.SetString(fmt.Sprintf("%s %s", helpKey.Key, helpKey.Desc))
	}

	padding := 1
	footer := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftSection.Render(),
		lipgloss.PlaceHorizontal(
			m.ctx.ScreenWidth-lipgloss.Width(leftSection.String())-2*padding,
			lipgloss.Right,
			rightSection.Render(),
		),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Width(m.ctx.ScreenWidth).
		PaddingTop(m.ctx.ScreenHeight - m.ctx.ContentHeight - lipgloss.Height(footer)).
		PaddingLeft(padding).
		PaddingRight(padding).
		Render(footer)
}
