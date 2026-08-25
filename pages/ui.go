package pages

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/nav"
	"go.dalton.dog/bubbleup"
)

type Model struct {
	alert  bubbleup.AlertModel
	footer FooterComponents
}

func NewUI(appVersion string) Model {
	return Model{
		alert:  newAlertModel(80),
		footer: NewFooter(appVersion),
	}
}

func newAlertModel(width int) bubbleup.AlertModel {
	alert := bubbleup.NewAlertModel(max(width, 20), false)
	registerDefaultAlertTypes(alert)
	return *alert
}

func (m Model) Init() tea.Cmd {

	return tea.Sequence(
		// 跳转到开平页面
		m.footer.Init(),
		m.alert.Init(),
		// 跳转去开屏页面
		nav.Push(newSplashPage()),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmds     []tea.Cmd
		c        tea.Cmd
		alertCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g.Window.SetSize(msg)
		m.alert = newAlertModel(msg.Width - 2)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.Space):
			cmds = append(cmds, nav.PushOrBack(newBossPage()))
		case key.Matches(msg, consts.AppKeyMap.HelpPage):
			cmds = append(cmds, nav.PushOrBack(newHelpPage()))
		case key.Matches(msg, consts.AppKeyMap.SettingPage):
			// Comma opens settings from regular pages, but remains available as
			// the node separator while editing the settings form.
			if nav.IsActivePage(newSettingPage()) {
				if msg.String() == "," {
					break
				}
				return m, nav.Back()
			}
			cmds = append(cmds, nav.PushOrBack(newSettingPage()))
		case key.Matches(msg, consts.AppKeyMap.KeyQ):
			if nav.IsActivePage(newTopicPage()) {
				return m, tea.Quit
			}
			return m, nav.Back()
		case key.Matches(msg, consts.AppKeyMap.CtrlQuit):
			return m, tea.Quit
		}
	}

	// 警告框
	cmds = append(cmds, alertCmd)
	alert, c := m.alert.Update(msg)
	cmds = append(cmds, c)
	if a, ok := alert.(bubbleup.AlertModel); ok {
		m.alert = a
	}

	// 底部内容更新
	m.footer, c = m.footer.Update(msg)
	cmds = append(cmds, c)
	// 路由控制
	cmds = append(cmds, nav.Update(msg))

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	var (
		_, h = g.Window.GetSize()
	)

	var (
		footer       = m.footer.View()
		footerHeight = lipgloss.Height(footer)
		body         = lipgloss.
				NewStyle().
				Height(max(h-footerHeight, 1)).
				Render(nav.View())
		content = lipgloss.JoinVertical(
			lipgloss.Top,
			body,
			footer,
		)
	)

	return m.alert.Render(content)
}
