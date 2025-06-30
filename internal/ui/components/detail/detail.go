package detail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/dromara/carbon/v2"
	"github.com/muesli/reflow/wrap"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/model/response"
	"github.com/seth-shi/go-v2ex/internal/pkg"
	"github.com/seth-shi/go-v2ex/internal/ui/styles"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
)

const (
	keyHelp = "[q:返回 e:加载评论 r:加载图片 w/s/鼠标:滑动 a/d:翻页 =:显示页脚]"
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
	viewport     viewport.Model
	canLoadReply bool

	content      bytes.Buffer
	imageDataMap map[string]string

	id         int64
	replyPage  int
	replyIndex int
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
		m.id = msgType.ID
		m.replyPage = 0
		m.replyIndex = 0
		m.canLoadReply = true
		m.content.Reset()
		m.imageDataMap = make(map[string]string)
		m.viewport = viewport.New(config.Screen.Width-2, config.Screen.Height-lipgloss.Height(m.headerView())-2)
		return m, m.getDetail(msgType.ID)
	case messages.GetDetailResponse:
		cmds = append(cmds, m.renderDetail(msgType.Data))
	case messages.GetReplyResponse:
		cmds = append(cmds, m.onReplyResult(msgType))
		// 图片加载成功
	case messages.GetImageRequest:
		if messages.LoadingRequestImage.Loading() {
			return m, messages.Post(errors.New("请求图片中"))
		}
		return m, tea.Sequence(
			messages.LoadingRequestImage.PostStart(),
			m.requestImages(msgType.URL),
			messages.LoadingRequestImage.PostEnd(),
		)
	case messages.GetImageResult:
		cmds = append(cmds, m.onImageLoaded(msgType))
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
		case key.Matches(msgType, consts.AppKeyMap.LoadImage):
			return m, messages.Post(
				messages.GetImageRequest{URL: pkg.ExtractImgURLs(m.content.String())},
			)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	if m.content.Len() == 0 {
		return styles.Hint.
			Width(config.Screen.Width).
			PaddingTop(2).
			Bold(true).
			Height(1).
			Align(lipgloss.Center).
			Render("载入中...")
	}

	return fmt.Sprintf("%s\n%s", m.headerView(), m.viewport.View())
}

func (m Model) headerView() string {
	var p = 0.0
	if m.content.Len() > 0 {
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
	m.replyPage = msgType.CurrPage
	var cmds []tea.Cmd
	if data.Pagination.TotalCount > 0 {
		cmds = append(
			cmds, messages.Post(
				messages.ShowAlertRequest{
					Text: data.Pagination.ToString(m.replyPage),
					Help: keyHelp,
				},
			),
		)
	}

	if m.replyPage >= data.Pagination.TotalPages {
		m.canLoadReply = false
	}

	var (
		replies strings.Builder
		// 第一页评论展示顶部
		// 最后一页展示底部边框
		boxStyle = styles.
				Style.
				BorderTop(msgType.CurrPage == 1).
				BorderBottom(msgType.CurrPage == data.Pagination.TotalPages)
		replyTitleStyle = styles.
				Border.
				Width(config.Screen.GetContentWidth()).
				BorderRight(false).
				BorderBottom(false)
	)

	for _, r := range data.Result {
		m.replyIndex++
		floor := fmt.Sprintf(
			"#%d · %s @%s",
			m.replyIndex,
			carbon.CreateFromTimestamp(r.Created),
			r.Member.Username,
		)

		replies.WriteString(replyTitleStyle.Render(floor))
		replies.WriteString("\n")
		replies.WriteString(r.GetContent())
		replies.WriteString("\n\n")
	}

	// 这里处理图片替换
	m.content.WriteString(boxStyle.Width(config.Screen.Width).Render(replies.String()))
	m.refreshViewContent()
	return nil
}

func (m *Model) refreshViewContent() {

	var (
		imageStyle = styles.
			Border.
			BorderLeft(true).
			BorderRight(false).
			BorderTop(false).
			BorderBottom(false).
			PaddingLeft(1)
	)

	if len(m.imageDataMap) > 0 {
		var items []string
		var index int
		for k, v := range m.imageDataMap {
			index++

			var imageData strings.Builder
			imageData.WriteString(fmt.Sprintf("图片#%d", index))
			imageData.WriteString("\n")
			imageData.WriteString(v)
			items = append(items, k, fmt.Sprintf("\n%s", imageStyle.Render(imageData.String())))
		}
		replacer := strings.NewReplacer(items...)
		m.viewport.SetContent(replacer.Replace(m.content.String()))
		return
	}

	m.viewport.SetContent(m.content.String())
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

func (m *Model) requestImages(urls []string) tea.Cmd {

	return func() tea.Msg {

		if len(urls) == 0 {
			return errors.New("当前页面无图片")
		}

		width := (config.Screen.Width * 9) / 10
		return messages.GetImageResult{
			Result: pkg.ProcessURLs(urls, width),
		}
	}
}
func (m *Model) onImageLoaded(result messages.GetImageResult) tea.Cmd {
	m.imageDataMap = lo.Assign(m.imageDataMap, result.Result)
	m.refreshViewContent()
	return nil
}

func (m *Model) renderDetail(detail response.V2DetailResult) tea.Cmd {

	var (
		contentWidth      = config.Screen.Width - 2
		content           strings.Builder
		topicContent      = detail.GetContent()
		contentTitleStyle = styles.Border.BorderRight(false).BorderBottom(false)
	)

	content.WriteString(
		contentTitleStyle.
			Width(config.Screen.Width).
			Render(
				fmt.Sprintf(
					"V2EX > %s %s\n%s · %s · %d 回复\n\n%s\n\n%s",
					styles.Bold.Render(detail.Node.Title),
					detail.Url,
					detail.Member.Username, carbon.CreateFromTimestamp(detail.Created),
					detail.Replies,
					lipgloss.NewStyle().
						Bold(true).
						Border(lipgloss.RoundedBorder(), false, false, true, false).
						Render(detail.Title),
					wrap.String(topicContent, contentWidth),
				),
			),
	)
	content.WriteString("\n\n")

	// 附言
	for i, c := range detail.Supplements {

		desc := fmt.Sprintf(
			"第 %d 条附言 · %s\n%s", i+1, carbon.CreateFromTimestamp(c.Created),
			c.GetContent(),
		)
		content.WriteString(contentTitleStyle.Width(config.Screen.Width).Render(desc))
	}

	// 构建内容
	m.content.WriteString(content.String())
	m.refreshViewContent()
	// 如果有评论去加载评论
	if detail.Replies > 0 {
		return m.getReply(detail.Id)
	}

	return messages.Post(messages.ShowToastRequest{Text: "无评论加载"})
}
