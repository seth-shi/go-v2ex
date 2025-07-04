package pages

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/ui/styles"
)

type helpPage struct {
	keys consts.KeyMap
	help help.Model
	windowPage
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
	return nil
}

func (m helpPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	m.windowPage = m.windowPage.Update(msg)

	return m, nil
}

func (m helpPage) View() string {

	more := styles.Bold.Render("\n如有请求超时, 请设置 clash 全局代理, 或者复制代理环境变量到终端执行")

	return lipgloss.
		NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(m.w - 2).
		Render(lipgloss.JoinVertical(lipgloss.Top, m.help.View(m.keys), more))
}
