package pages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/dromara/carbon/v2"
	"github.com/muesli/reflow/wrap"
	"github.com/pkg/browser"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/api"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/pkg"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/seth-shi/go-v2ex/v2/styles"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/messages"
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

type detailPage struct {
	url          string
	viewport     viewport.Model
	canLoadReply bool

	content      bytes.Buffer
	imageDataMap map[string]string

	id         int64
	replyPage  int
	replyIndex int
	opMember   response.MemberResult
}

func newDetailPage() detailPage {
	return detailPage{}
}

func (m detailPage) Init() tea.Cmd {
	return nil
}

func (m detailPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// 显示是否是 pro && 是否有人

	switch msg := msg.(type) {
	case messages.GetDetailRequest:
		// 获取内容 + 第一页的评论
		var (
			w, h = g.Window.GetSize()
		)
		m.id = msg.ID
		m.replyPage = 0
		m.replyIndex = 0
		m.canLoadReply = true
		m.content.Reset()
		m.imageDataMap = make(map[string]string)
		m.viewport = viewport.New(w-2, h-lipgloss.Height(m.headerView())-2)
		return m, m.getDetail(msg.ID)
	case messages.GetDetailResponse:
		cmds = append(cmds, m.onDetailResult(msg.Data))
	case messages.GetReplyResponse:
		cmds = append(cmds, m.onReplyResult(msg))
		// 图片加载成功
	case messages.GetImageRequest:
		if messages.LoadingRequestImage.Loading() {
			return m, commands.Post(errors.New("请求图片中"))
		}
		return m, tea.Sequence(
			messages.LoadingRequestImage.PostStart(),
			m.requestImages(msg.URL),
			messages.LoadingRequestImage.PostEnd(),
		)
	case messages.GetImageResult:
		cmds = append(cmds, m.onImageLoaded(msg))
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.KeyE):
			return m, m.getReply(m.id)
		case key.Matches(msg, consts.AppKeyMap.KeyQ):
			return m, commands.RedirectPop()
		case key.Matches(msg, consts.AppKeyMap.Up):
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case key.Matches(msg, consts.AppKeyMap.Down):
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case key.Matches(msg, consts.AppKeyMap.Left):
			msg = tea.KeyMsg{Type: tea.KeyPgUp}
		case key.Matches(msg, consts.AppKeyMap.Right):
			msg = tea.KeyMsg{Type: tea.KeyPgDown}
		case key.Matches(msg, consts.AppKeyMap.KeyR):
			return m, commands.Post(
				messages.GetImageRequest{URL: pkg.ExtractImgURLs(m.content.String())},
			)
		case key.Matches(msg, consts.AppKeyMap.F1):
			return m, func() tea.Msg {
				return browser.OpenURL(m.url)
			}
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m detailPage) View() string {

	if m.content.Len() == 0 {
		return loadingView("正在加载内容...")
	}

	return fmt.Sprintf("%s\n%s", m.headerView(), m.viewport.View())
}

func (m detailPage) headerView() string {
	var p = 0.0
	if m.content.Len() > 0 {
		p = m.viewport.ScrollPercent() * 100
	}
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", p))
	line := strings.Repeat("─", max(0, int(math.Ceil(float64(m.viewport.Width-lipgloss.Width(info))*p/100))))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m *detailPage) getReply(id int64) tea.Cmd {

	if messages.LoadingRequestReply.Loading() {
		return commands.Post(errors.New("评论请求中"))
	}

	if !m.canLoadReply {
		return commands.Post(errors.New("已无更多评论"))
	}

	return tea.Sequence(
		messages.LoadingRequestReply.PostStart(),
		api.V2ex.GetReply(context.Background(), id, m.replyPage+1),
		messages.LoadingRequestReply.PostEnd(),
	)
}

func (m *detailPage) onReplyResult(msgType messages.GetReplyResponse) tea.Cmd {

	data := msgType.Data
	//  请求之后增加分页, 防止网络失败, 增加了分页
	m.replyPage = msgType.CurrPage
	if m.replyPage >= data.Pagination.TotalPages {
		m.canLoadReply = false
	}

	var (
		pageInfo = data.Pagination.ToString(m.replyPage)
		w, _     = g.Window.GetSize()
		replies  strings.Builder
		// 第一页评论展示顶部
		// 最后一页展示底部边框
		boxStyle = styles.
				Style.
				BorderTop(msgType.CurrPage == 1).
				BorderBottom(msgType.CurrPage == data.Pagination.TotalPages)
		replyTitleStyle = styles.
				Border.
				Width(w - 2).
				BorderRight(false).
				BorderBottom(false)
		me = g.Me.Get()
	)

	replies.WriteString("\n")

	for _, r := range data.Result {
		m.replyIndex++

		var (
			opText string
		)

		if r.Member.Id == m.opMember.Id {
			opText = styles.MemberOp
		}

		floor := fmt.Sprintf(
			"#%d · %s @%s%s",
			m.replyIndex,
			carbon.CreateFromTimestamp(r.Created),
			r.Member.GetUserNameLabel(me.Id),
			opText,
		)

		replies.WriteString(replyTitleStyle.Render(floor))
		replies.WriteString("\n")
		replies.WriteString(r.GetContent(w))
		replies.WriteString("\n")
	}

	// 这里处理图片替换
	m.content.WriteString(boxStyle.Width(w).Render(replies.String()))
	m.refreshViewContent()
	return m.showStatusBarText(pageInfo)
}

func (m *detailPage) refreshViewContent() {

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

func (m *detailPage) requestImages(urls []string) tea.Cmd {

	return func() tea.Msg {

		if len(urls) == 0 {
			return errors.New("当前页面无图片")
		}

		// 只去下载图片里没有的
		keys := lo.Keys(m.imageDataMap)
		diffUrl := lo.Without(urls, keys...)

		slog.Info("下载图片", slog.Int("count", len(diffUrl)))

		var (
			w, _ = g.Window.GetSize()
		)

		width := (w * 9) / 10
		return messages.GetImageResult{
			Result: pkg.ProcessURLs(diffUrl, width),
		}
	}
}
func (m *detailPage) onImageLoaded(result messages.GetImageResult) tea.Cmd {
	m.imageDataMap = lo.Assign(m.imageDataMap, result.Result)
	m.refreshViewContent()
	return nil
}

func (m detailPage) getDetail(id int64) tea.Cmd {
	return tea.Sequence(
		messages.LoadingRequestDetail.PostStart(),
		api.V2ex.GetDetail(context.Background(), id),
		messages.LoadingRequestDetail.PostEnd(),
	)
}
func (m *detailPage) onDetailResult(detail response.V2DetailResult) tea.Cmd {

	var (
		w, _              = g.Window.GetSize()
		contentWidth      = w - 2
		content           strings.Builder
		topicContent      = detail.GetContent(w)
		contentTitleStyle = styles.Border.BorderRight(false).BorderBottom(false)
		me                = g.Me.Get()
	)
	m.opMember = detail.Member
	m.url = detail.Url

	content.WriteString(
		contentTitleStyle.
			Width(w).
			Render(
				fmt.Sprintf(
					"V2EX > %s %s\n\n%s · %s · %d 回复\n\n%s\n\n%s",
					styles.Bold.Render(detail.Node.Title),
					detail.Url,
					detail.Member.GetUserNameLabel(me.Id),
					carbon.CreateFromTimestamp(detail.Created),
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
			"#%d 条附言 · %s\n%s", i+1, carbon.CreateFromTimestamp(c.Created),
			c.GetContent(w),
		)
		content.WriteString(contentTitleStyle.Width(w).Render(desc))
	}

	// 构建内容
	m.content.WriteString(content.String())
	m.refreshViewContent()
	// 如果有评论去加载评论
	if detail.Replies > 0 {
		return m.getReply(detail.Id)
	}

	m.canLoadReply = false
	return m.showStatusBarText("0 回复")
}

func (m detailPage) showStatusBarText(firstText string) tea.Cmd {
	return commands.Post(messages.ShowStatusBarTextRequest{FirstText: firstText})
}
