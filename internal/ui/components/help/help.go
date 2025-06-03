package help

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/consts"
)

type Model struct {
	keys consts.KeyMap
	help help.Model
}

func New() Model {
	helpModel := help.New()
	helpModel.ShowAll = true
	m := Model{
		help: helpModel,
		keys: consts.AppKeyMap,
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {

	more := "\n如有请求超时, 请设置 clash 全局代理, 或者复制代理环境变量到终端执行"

	return lipgloss.
		NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(config.Screen.Width - 2).
		Height(config.Screen.Height - 4).
		Padding(1).
		Render(lipgloss.JoinVertical(lipgloss.Top, m.help.View(m.keys), more))
}
