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
		m.viewport = viewport.New(config.Screen.Width-2, config.Screen.Height-lipgloss.Height(m.headerView())+2)
		// 开启定时器去获取评论列表
		return m, tea.Batch(messages.Post(messages.ShowTipsRequest{Text: "按 n 更多评论 ←返回列表"}), m.getDetail(msgType.ID), m.getReply(msgType.ID))
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
		messages.Post(messages.LoadingRequestDetail.Start), api.Client.GetDetail(id), messages.Post(messages.LoadingRequestDetail.End),
	)
}

func (m *Model) onReplyResult(msgType messages.GetRepliesResult) tea.Cmd {

	m.requestingReply = false
	if msgType.Error != nil {
		return messages.Post(msgType.Error)
	}

	// 如果返回的评论列表是空的, 那么就不再请求
	m.replyPage++
	m.replies = append(m.replies, msgType.Replies...)
	if len(msgType.Replies) == 0 {
		m.canRequestReply = false
		return messages.Post(messages.ShowAutoTipsRequest{Text: "已无更多评论"})
	}

	if len(msgType.Replies) < 20 {
		m.canRequestReply = false
		return messages.Post(messages.ShowAutoTipsRequest{Text: "已无更多评论"})
	}

	return nil
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
		messages.Post(messages.LoadingRequestReply.Start), api.Client.GetReply(id, m.replyPage), messages.Post(messages.LoadingRequestReply.End),
	)
}

var (
	descStyle = lipgloss.NewStyle().Underline(true)
)

func (m *Model) initViewport() {
	// 获取详情
	var (
		contentWidth = config.Screen.Width - 2
		content      strings.Builder
	)
	// 组装文案
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("V2EX > %s https://www.v2ex.com/t/%d", m.detail.Node.Title, m.detail.Id)))
	content.WriteString("\n\n")
	content.WriteString(descStyle.Render(fmt.Sprintf("%s · %s · %d 回复", m.detail.Member.Username, carbon.CreateFromTimestamp(m.detail.Created), m.detail.Replies)))
	content.WriteString("\n\n")
	content.WriteString(wrap.String(m.detail.GetContent(), contentWidth))
	content.WriteString("\n\n")

	// 附言
	for i, c := range m.detail.Supplements {
		content.WriteString(descStyle.Render(fmt.Sprintf("第 %d 条附言 · %s", i+1, carbon.CreateFromTimestamp(c.Created))))
		content.WriteString("\n")
		content.WriteString(c.GetContent())
		content.WriteString("\n\n")
	}

	// 开始渲染评论
	content.WriteString("\n\n")
	content.WriteString(descStyle.Bold(true).Render("回复列表"))
	content.WriteString("\n\n")
	for i, r := range m.replies {
		content.WriteString(descStyle.Render(fmt.Sprintf("#%d · %s @%s", i+1, carbon.CreateFromTimestamp(r.Created), r.Member.Username)))
		content.WriteString("\n")
		content.WriteString(r.GetContent())
		content.WriteString("\n\n")
	}

	m.viewport.SetContent(content.String())
	m.viewportReady = true
}
