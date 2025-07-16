package pages

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevm/bubbleo/navstack"
	"github.com/kevm/bubbleo/window"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"go.dalton.dog/bubbleup"
)

type Model struct {
	navigator *navstack.Model

	alert  bubbleup.AlertModel
	footer FooterComponents
}

func NewUI(appVersion string) Model {
	w := window.New(120, 30, 0, 0)
	ns := navstack.New(&w)

	alert := bubbleup.NewAlertModel(80, false)
	registerDefaultAlertTypes(alert)
	return Model{
		navigator: &ns,
		alert:     lo.FromPtr(alert),
		footer:    NewFooter(appVersion),
	}
}

func (m Model) Init() tea.Cmd {

	return tea.Sequence(
		// 跳转到开平页面
		m.footer.Init(),
		m.alert.Init(),
		// 跳转去开屏页面
		commands.Redirect(RouteSplash),
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
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.Space):
			cmds = append(cmds, redirectIfSamePop(m.navigator.Top(), RouteBoss))
		case key.Matches(msg, consts.AppKeyMap.HelpPage):
			cmds = append(cmds, redirectIfSamePop(m.navigator.Top(), RouteHelp))
		case key.Matches(msg, consts.AppKeyMap.SettingPage):
			cmds = append(cmds, redirectIfSamePop(m.navigator.Top(), RouteSetting))
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
	cmds = append(cmds, m.navigator.Update(msg))

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	var (
		_, h = g.Window.GetSize()
		sess = g.Session
	)

	var (
		footer       = m.footer.View()
		footerHeight = lipgloss.Height(footer)
		body         = lipgloss.
				NewStyle().
				Height(h - footerHeight).
				Render(m.navigator.View())
		content = lipgloss.JoinVertical(
			lipgloss.Top,
			body,
			footer,
		)
	)

	if !sess.HideFooter.Load() {
		content = m.alert.Render(content)
	}

	return content
}
