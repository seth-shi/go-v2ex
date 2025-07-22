package pages

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path"
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
	id        int64
	viewport  viewport.Model
	decodeMap map[string]string

	replyPageInfo response.V2PageResponse
	contentDetail response.V2DetailResult
	contentReply  []response.V2ReplyResult
}

func newDetailPage() detailPage {
	return detailPage{
		decodeMap: make(map[string]string),
		viewport:  viewport.New(40, 40),
	}
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
		m.viewport.Width = w
		m.viewport.Height = h
		// 重新修改键盘映射
		m.viewport.KeyMap.Up = consts.AppKeyMap.Up
		m.viewport.KeyMap.Down = consts.AppKeyMap.Down
		m.viewport.KeyMap.Left = consts.AppKeyMap.Left
		m.viewport.KeyMap.Right = consts.AppKeyMap.Right
		return m, tea.Batch(m.getDetail(), m.getReply())
	case messages.GetDetailResponse:
		m.contentDetail = msg.Data
		cmds = append(cmds, m.renderContent())
	case messages.GetReplyResponse:
		m.replyPageInfo = msg.Data.Pagination
		m.contentReply = append(m.contentReply, msg.Data.Result...)
		cmds = append(cmds, m.renderContent())
	case messages.DecodeDetailContentResult:
		m.decodeMap = lo.Assign(m.decodeMap, msg.Result)
		cmds = append(cmds, m.renderContent())
	case messages.RenderDetailContentResult:
		m.viewport.SetContent(msg.Content)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.KeyE):
			if m.replyPageInfo.CurrPage >= m.replyPageInfo.TotalPages {
				return m, commands.Post(errors.New("已无更多评论"))
			}
			return m, m.getReply()
		case key.Matches(msg, consts.AppKeyMap.KeyR):
			return m.decodeContent()
		case key.Matches(msg, consts.AppKeyMap.F1):
			return m, func() tea.Msg {
				return browser.OpenURL(m.contentDetail.Url)
			}
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m detailPage) View() string {
	return fmt.Sprintf("%s\n%s", m.headerView(), m.viewport.View())
}

func (m detailPage) headerView() string {
	var p = 0.0
	if m.contentDetail.Id > 0 {
		p = m.viewport.ScrollPercent() * 100
	}
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", p))
	line := strings.Repeat("─", max(0, int(math.Ceil(float64(m.viewport.Width-lipgloss.Width(info))*p/100))))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m detailPage) getReply() tea.Cmd {
	// 默认第 0 页
	page := m.replyPageInfo.CurrPage + 1
	return commands.
		LoadingRequestReply.
		Run(api.V2ex.GetReply(context.Background(), m.id, page))
}

func (m detailPage) renderContent() tea.Cmd {

	return tea.Batch(
		func() tea.Msg {
			return messages.RenderDetailContentResult{Content: m.buildContent()}
		},
		commands.Post(messages.ShowStatusBarTextRequest{FirstText: m.replyPageInfo.ToString("0 评论")}),
	)
}
func (m detailPage) buildContent() string {
	var (
		w, _              = g.Window.GetSize()
		content           strings.Builder
		contentTitleStyle = styles.Border.BorderRight(false).BorderBottom(false)
		// 第一页评论展示顶部
		// 最后一页展示底部边框
		boxStyle = styles.
				Style.
				BorderTop(true).
				BorderBottom(true)
		replyTitleStyle = styles.
				Border.
				Width(w - 2).
				BorderRight(false).
				BorderBottom(false)
		me          = g.Me.Get()
		decodeItems []string
	)
	for k, v := range m.decodeMap {
		decodeItems = append(decodeItems, k, v)
	}

	decodeFn := func(t string) string {
		if len(decodeItems) == 0 {
			return t
		}

		return strings.NewReplacer(decodeItems...).Replace(t)
	}

	// //////////////////////////////////////
	// 主题内容渲染
	content.WriteString(
		contentTitleStyle.
			Width(w).
			Render(
				fmt.Sprintf(
					"V2EX > %s %s\n\n%s · %s · %d 回复\n\n%s\n\n%s",
					styles.Bold.Render(m.contentDetail.Node.Title),
					m.contentDetail.Url,
					m.contentDetail.Member.GetUserNameLabel(me.Id),
					carbon.CreateFromTimestamp(m.contentDetail.Created),
					m.contentDetail.Replies,
					lipgloss.NewStyle().
						Bold(true).
						Border(lipgloss.RoundedBorder(), false, false, true, false).
						Render(m.contentDetail.Title),
					wrap.String(decodeFn(m.contentDetail.GetContent(w)), w-2),
				),
			),
	)
	content.WriteString("\n\n")

	// //////////////////////////////////////
	// 附言
	for i, c := range m.contentDetail.Supplements {

		desc := fmt.Sprintf(
			"#%d 条附言 · %s\n%s", i+1, carbon.CreateFromTimestamp(c.Created),
			decodeFn(c.GetContent(w)),
		)
		content.WriteString(contentTitleStyle.Width(w).Render(desc))
	}

	content.WriteString("\n")
	// //////////////////////////////////////
	// 评论的渲染列表
	var replyContent strings.Builder
	for i, r := range m.contentReply {
		var (
			// 是否是楼主
			opText = lo.If(r.Member.Id == m.contentDetail.Member.Id, styles.MemberOp).Else("")
		)

		floor := fmt.Sprintf(
			"#%d · %s @%s%s",
			i,
			carbon.CreateFromTimestamp(r.Created),
			r.Member.GetUserNameLabel(me.Id),
			opText,
		)
		replyContent.WriteString(replyTitleStyle.Render(floor))
		replyContent.WriteString("\n")
		replyContent.WriteString(decodeFn(r.GetContent(w)))
		replyContent.WriteString("\n")
	}
	content.WriteString(boxStyle.Width(w).Render(replyContent.String()))
	return content.String()
}

func (m detailPage) getDetail() tea.Cmd {
	return commands.
		LoadingRequestDetail.
		Run(api.V2ex.GetDetail(context.Background(), m.id))
}

func (m detailPage) decodeContent() (tea.Model, tea.Cmd) {

	// 两种方式接码
	cmd := commands.
		LoadingDecodeContent.
		Run(
			func() tea.Msg {

				var (
					content    = m.buildContent()
					w, _       = g.Window.GetSize()
					keys       = lo.Keys(m.decodeMap)
					urls       = pkg.ExtractImgURLs(content)
					diffUrl    = lo.Without(urls, keys...)
					width      = (w * 9) / 10
					replaceMap = make(map[string]string)
					imageStyle = styles.
							Border.
							BorderLeft(true).
							BorderRight(false).
							BorderTop(false).
							BorderBottom(false).
							PaddingLeft(1)
					tagStyle = styles.Active.Bold(true).Underline(true)
				)

				for k, v := range pkg.DownloadImageURL(diffUrl, width) {
					// 图片显示的格式: xxx => 换行: 替换#1 换行 图片
					index := path.Base(k)
					var imageData strings.Builder
					imageData.WriteString(tagStyle.Render(fmt.Sprintf("图片解码 #%s", index)))
					imageData.WriteString("\n")
					imageData.WriteString(v)
					replaceMap[k] = fmt.Sprintf("\n%s", imageStyle.Render(imageData.String()))
				}

				index := 0
				for k, v := range pkg.DetectBase64(content) {
					index++
					replaceMap[k] = tagStyle.Render(fmt.Sprintf("base64解码#%d %s", index, v))
				}

				if len(replaceMap) == 0 {
					return errors.New("无数据需要解码")
				}

				return messages.DecodeDetailContentResult{
					Result: replaceMap,
				}
			},
		)
	return m, cmd
}
