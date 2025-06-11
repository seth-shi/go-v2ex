package detail

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/model/response"

	"github.com/dromara/carbon/v2"
	"github.com/muesli/reflow/wrap"
	"github.com/seth-shi/go-v2ex/internal/api"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
)

const (
	keyHelp = "[q:返回 e:加载评论 w/s/鼠标:滑动 a/d:翻页 =:显示页脚]"
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
			Border(lipgloss.RoundedBorder())
)

type Model struct {
	viewport      viewport.Model
	viewportReady bool
	detail        response.V2DetailResult
	replies       []response.V2ReplyResult
	canLoadReply  bool

	id        int64
	replyPage int
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
		m.id = msgType.ID
		m.replyPage = 0
		m.canLoadReply = true
		m.viewport = viewport.New(config.Screen.Width-2, config.Screen.Height-lipgloss.Height(m.headerView())-2)
		return m, tea.Batch(
			m.getDetail(msgType.ID),
			m.getReply(msgType.ID),
		)
	case messages.GetDetailResponse:
		m.detail = msgType.Data
		m.initViewport()
	case messages.GetReplyResponse:
		cmds = append(cmds, m.onReplyResult(msgType))
		m.initViewport()
	case tea.WindowSizeMsg:
		m.initViewport()
	case tea.KeyMsg:
		switch {
		case key.Matches(msgType, consts.AppKeyMap.Enter):
			return m, m.getReply(m.id)
		// 回到首页
		case key.Matches(msgType, consts.AppKeyMap.Back):
			return m, messages.Post(messages.RedirectTopicsPage{})
		case key.Matches(msgType, consts.AppKeyMap.Up):
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case key.Matches(msgType, consts.AppKeyMap.Down):
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case key.Matches(msgType, consts.AppKeyMap.Left):
			msg = tea.KeyMsg{Type: tea.KeyPgUp}
		case key.Matches(msgType, consts.AppKeyMap.Right):
			msg = tea.KeyMsg{Type: tea.KeyPgDown}
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
		messages.LoadingRequestDetail.PostStart(),
		api.V2ex.GetDetail(context.Background(), id),
		messages.LoadingRequestDetail.PostEnd(),
	)
}

func (m *Model) onReplyResult(msgType messages.GetReplyResponse) tea.Cmd {

	data := msgType.Data
	//  请求之后增加分页, 防止网络失败, 增加了分页
	m.replies = append(m.replies, msgType.Data.Result...)
	m.replyPage = data.Pagination.CurrPage
	var cmds []tea.Cmd
	if data.Pagination.TotalCount > 0 {
		cmds = append(
			cmds, messages.Post(
				messages.ShowAlertRequest{
					Text: data.Pagination.ToString(),
					Help: keyHelp,
				},
			),
		)
	}

	if m.replyPage >= data.Pagination.TotalPages {
		m.canLoadReply = false
	}

	return tea.Batch(cmds...)
}

func (m *Model) getReply(id int64) tea.Cmd {

	if messages.LoadingRequestReply.Loading() {
		return messages.Post(errors.New("评论请求中"))
	}

	if !m.canLoadReply {
		return messages.Post(errors.New("已无更多评论"))
	}

	return tea.Sequence(
		messages.LoadingRequestReply.PostStart(),
		api.V2ex.GetReply(context.Background(), id, m.replyPage+1),
		messages.LoadingRequestReply.PostEnd(),
	)
}

func (m *Model) initViewport() {
	// 获取详情
	var (
		contentWidth = config.Screen.Width - 2
		content      strings.Builder
	)
	// 组装文案
	// 找到所有图片去动态替换成字符串
	content.WriteString(
		sectionStyle.
			Width(config.Screen.Width).
			Render(
				fmt.Sprintf(
					"V2EX > %s %s\n%s · %s · %d 回复\n\n%s\n\n%s",
					m.detail.Node.Title, m.detail.Url,
					m.detail.Member.Username, carbon.CreateFromTimestamp(m.detail.Created),
					m.detail.Replies,
					lipgloss.NewStyle().
						Bold(true).
						Border(lipgloss.RoundedBorder(), false, false, true, false).
						Render(m.detail.Title),
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
			floor := fmt.Sprintf(
				"#%d · %s @%s",
				i+1,
				carbon.CreateFromTimestamp(r.Created),
				r.Member.Username,
			)
			replies.WriteString(lipgloss.NewStyle().Bold(true).Render(floor))
			replies.WriteString("\n")
			replies.WriteString(r.GetContent())
			replies.WriteString("\n\n")
		}
		content.WriteString(sectionStyle.Width(config.Screen.Width).Render(replies.String()))
	}

	m.viewport.SetContent(content.String())
	m.viewportReady = true
}
