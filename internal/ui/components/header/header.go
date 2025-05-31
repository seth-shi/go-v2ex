package header

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

type Model struct {
	me     *types.V2MemberResult
	config types.FileConfig
	screen types.ScreenSize
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch typeMsg := msg.(type) {
	case messages.GetMeRequest:
		return m, tea.Batch(messages.Post(messages.LoadingRequestMe.Start), api.Client.GetMember)
	case messages.GetMeResult:
		m.me = typeMsg.Member
		return m, tea.Batch(messages.Post(messages.LoadingRequestMe.End), messages.Post(typeMsg.Error))
	case tea.WindowSizeMsg:
		m.screen.Sync(typeMsg)
		return m, nil
	}

	return m, nil
}
func (m Model) View() string {

	var (
		leftText  = ""
		rightText = "未登录"
	)

	if m.me != nil {
		rightText = m.me.Username
	} else if m.config.Token != "" {
		rightText = "登录中..."
	}

	padding := 1
	view := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		Width(m.screen.Width).
		PaddingLeft(1).
		PaddingRight(1).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				leftText,
				lipgloss.PlaceHorizontal(
					m.screen.Width-lipgloss.Width(leftText)-2*padding,
					lipgloss.Right,
					rightText,
				),
			),
		)
	return view
}
