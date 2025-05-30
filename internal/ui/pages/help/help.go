package help

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/ui/context"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

type Model struct {
	keys consts.KeyMap
	help help.Model
}

func New(ctx *context.Data) Model {
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

	switch typeMsg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(typeMsg, consts.AppKeyMap.Back):
			return m, func() tea.Msg {
				return messages.GoToHome{}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	more := "\n如有请求超时, 请设置 clash 全局代理, 或者复制代理环境变量到终端执行"

	return lipgloss.
		NewStyle().
		Padding(1).
		Render(lipgloss.JoinVertical(lipgloss.Top, m.help.View(m.keys), more))
}
