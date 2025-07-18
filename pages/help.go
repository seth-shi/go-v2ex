package pages

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

type helpPage struct {
	keys consts.KeyMap
	help help.Model
}

func newHelpPage() helpPage {
	helpModel := help.New()
	helpModel.ShowAll = true
	m := helpPage{
		help: helpModel,
		keys: consts.AppKeyMap,
	}

	return m
}

func (m helpPage) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			g.Session.HideFooter.Store(true)
			return nil
		},
	)
}

func (m helpPage) Close() error {
	g.Session.HideFooter.Store(false)
	return nil
}

func (m helpPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.KeyQ):
			return m, commands.RedirectPop()
		}
	}

	return m, nil
}

func (m helpPage) View() string {

	var (
		w, _ = g.Window.GetSize()
	)
	more := styles.Bold.Render("\n如有请求超时, 请设置 clash 全局代理, 或者复制代理环境变量到终端执行")
	return lipgloss.
		NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(w - 2).
		Render(lipgloss.JoinVertical(lipgloss.Top, m.help.View(m.keys), more))
}
