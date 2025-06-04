package detail

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/seth-shi/go-v2ex/internal/consts"

	"github.com/dromara/carbon/v2"
	"github.com/muesli/reflow/wrap"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/types"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

const (
	keyHelp = "[n ←]"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
	sectionStyle = lipgloss.
			NewStyle().
			Border(lipgloss.RoundedBorder()).
			Bold(true)
)

type Model struct {
	viewport        viewport.Model
	viewportReady   bool
	detail          types.V2DetailResult
	replies         []types.V2ReplyResult
	canRequestReply bool

	id              int64
	replyPage       int
	requestingReply bool
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msgType := msg.(type) {
	case messages.GetDetailRequest:
		// 获取内容 + 第一页的评论
		m.viewportReady = false
		m.canRequestReply = true
		m.id = msgType.ID
		m.replyPage = 1
		m.viewport = viewport.New(config.Screen.Width-2, config.Screen.Height-lipgloss.Height(m.headerView())-2)
		return m, tea.Batch(
			messages.Post(messages.ShowTipsRequest{Text: keyHelp}), m.getDetail(msgType.ID),
			m.getReply(msgType.ID),
		)
	case messages.GetDetailResult:
		m.detail = msgType.Detail
		m.initViewport()
	case messages.GetRepliesResult:
		cmds = append(cmds, m.onReplyResult(msgType))
		m.initViewport()
	case tea.WindowSizeMsg:
		m.initViewport()
	case tea.KeyMsg:
		switch {
		// 回到首页
		case msgType.String() == "n":
			return m, m.getReply(m.id)
		case key.Matches(msgType, consts.AppKeyMap.Left):
			return m, messages.Post(messages.RedirectTopicsPage{})
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return fmt.Sprintf("%s\n%s", m.headerView(), m.viewport.View())
}

func (m Model) headerView() string {
	var p = 0.0
	if m.viewportReady {
		p = m.viewport.ScrollPercent() * 100
	}
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", p))
	line := strings.Repeat("─", max(0, int(math.Ceil(float64(m.viewport.Width-lipgloss.Width(info))*p/100))))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m Model) getDetail(id int64) tea.Cmd {
	return tea.Sequence(
		messages.Post(messages.LoadingRequestDetail.Start), api.Client.GetDetail(id),
		messages.Post(messages.LoadingRequestDetail.End),
	)
}

func (m *Model) onReplyResult(msgType messages.GetRepliesResult) tea.Cmd {

	m.requestingReply = false
	if msgType.Error != nil {
		return messages.Post(msgType.Error)
	}

	//  请求之后增加分页, 防止网络失败, 增加了分页
	m.replyPage++
	m.replies = append(m.replies, msgType.Replies...)

	var cmds []tea.Cmd

	if msgType.Pagination.Total > 0 {
		cmds = append(
			cmds, messages.Post(
				messages.ShowTipsRequest{
					Text: msgType.Pagination.ToString(keyHelp),
				},
			),
		)
	}

	if m.replyPage > msgType.Pagination.Pages {
		m.canRequestReply = false
		cmds = append(cmds, messages.Post(messages.ShowTipsRequest{Text: "没有更多了"}))
	}

	return tea.Batch(cmds...)
}

func (m *Model) getReply(id int64) tea.Cmd {

	if m.requestingReply {
		return messages.Post(errors.New("评论请求中"))
	}

	if !m.canRequestReply {
		return nil
	}
	m.requestingReply = true
	return tea.Sequence(
		messages.Post(messages.LoadingRequestReply.Start), api.Client.GetReply(id, m.replyPage),
		messages.Post(messages.LoadingRequestReply.End),
	)
}

func (m *Model) initViewport() {
	// 获取详情
	var (
		contentWidth = config.Screen.Width - 2
		content      strings.Builder
	)
	// 组装文案
	content.WriteString(
		sectionStyle.
			Width(config.Screen.Width).
			Render(
				fmt.Sprintf(
					"V2EX > %s %s\n%s · %s · %d 回复\n%s",
					m.detail.Node.Title, m.detail.Url,
					m.detail.Member.Username, carbon.CreateFromTimestamp(m.detail.Created),
					m.detail.Replies,
					wrap.String(m.detail.GetContent(), contentWidth),
				),
			),
	)
	content.WriteString("\n\n")

	// 附言
	for i, c := range m.detail.Supplements {

		desc := fmt.Sprintf(
			"第 %d 条附言 · %s\n%s", i+1, carbon.CreateFromTimestamp(c.Created),
			c.GetContent(),
		)
		content.WriteString(sectionStyle.Width(config.Screen.Width).Render(desc))
	}

	// 开始渲染评论
	content.WriteString("\n\n")
	if len(m.replies) > 0 {
		var replies strings.Builder
		for i, r := range m.replies {
			replies.WriteString(
				fmt.Sprintf(
					"#%d · %s @%s", i+1, carbon.CreateFromTimestamp(r.Created), r.Member.Username,
				),
			)
			replies.WriteString("\n")
			replies.WriteString(r.GetContent())
			replies.WriteString("\n\n")
		}
		content.WriteString(sectionStyle.Width(config.Screen.Width).Render(replies.String()))
	}

	m.viewport.SetContent(content.String())
	m.viewportReady = true
}
