package header

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

type Model struct {
	me        *types.V2MemberResult
	rightText *string
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msgType := msg.(type) {
	case messages.GetMeRequest:
		m.rightText = lo.ToPtr("登录中...")
		return m, tea.Batch(messages.Post(messages.LoadingRequestMe.Start), api.Client.GetMember)
	case messages.GetMeResult:
		m.me = msgType.Member
		m.rightText = nil
		if msgType.Error != nil {
			m.rightText = lo.ToPtr("登录失败")
		}
		return m, tea.Batch(messages.Post(messages.LoadingRequestMe.End), messages.Post(msgType.Error))
	}

	return m, nil
}
func (m Model) View() string {

	padding := 1
	view := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		Width(config.Screen.Width).
		PaddingLeft(1).
		PaddingRight(1)
	if config.G.ShowHeader {

		var (
			leftText  = ""
			rightText = lo.FromPtr(m.rightText)
		)

		if m.me != nil {
			rightText = m.me.Username
		}

		view = view.SetString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftText,
			lipgloss.PlaceHorizontal(
				config.Screen.Width-lipgloss.Width(leftText)-2*padding,
				lipgloss.Right,
				rightText,
			),
		))
	}

	return view.String()
}
