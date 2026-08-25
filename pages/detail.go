package pages

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/pkg/browser"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/api"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/nav"
	"github.com/seth-shi/go-v2ex/v2/pkg"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/seth-shi/go-v2ex/v2/styles"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/messages"
)

var (
	_ nav.PageLife = detailPage{}
)

type detailPage struct {
	id        int64
	viewport  viewport.Model
	decodeMap map[string]string

	replyPageInfo response.V2PageResponse
	contentDetail response.V2DetailResult
	contentReply  []response.V2ReplyResult
}

func (m detailPage) OnEntering() (tea.Model, tea.Cmd) {
	// 获取内容 + 第一页的评论
	var (
		w, h = g.Window.GetSize()
	)
	m.viewport.Width = w
	m.viewport.Height = max(h-lipgloss.Height(m.headerView())-1, 1)
	// 重新修改键盘映射
	m.viewport.KeyMap.Up = consts.AppKeyMap.Up
	m.viewport.KeyMap.Down = consts.AppKeyMap.Down
	m.viewport.KeyMap.Left = consts.AppKeyMap.Left
	m.viewport.KeyMap.Right = consts.AppKeyMap.Right
	return m, nil
}

func (m detailPage) OnLeaving() (tea.Model, tea.Cmd) {
	return m, nil
}

func newDetailPage(id int64) detailPage {
	return detailPage{
		id:        id,
		decodeMap: make(map[string]string),
		viewport:  viewport.New(40, 40),
	}
}

func (m detailPage) Init() tea.Cmd {
	return tea.Batch(
		commands.
			LoadingRequestDetail.
			Run(api.V2ex.GetDetail(context.Background(), m.id)),
	)
}

func (m detailPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = max(msg.Width, 1)
		m.viewport.Height = max(msg.Height-lipgloss.Height(m.headerView())-1, 1)
		if m.contentDetail.Id > 0 {
			cmds = append(cmds, m.renderContent())
		}
	case messages.GetDetailResponse:
		m.contentDetail = msg.Data
		cmds = append(cmds, m.renderContent())
		if m.contentDetail.Replies > 0 {
			cmds = append(cmds, m.getReply())
		}
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
			if m.contentDetail.Replies == 0 || (m.replyPageInfo.TotalPages > 0 && m.replyPageInfo.CurrPage >= m.replyPageInfo.TotalPages) {
				return m, commands.Post(errors.New("已无更多评论"))
			}
			if m.replyPageInfo.CurrPage == 0 {
				return m, commands.Post(errors.New("评论正在加载中"))
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

	if m.contentDetail.Id == 0 {
		return loadingView(fmt.Sprintf("[%d]主题正在加载中...", m.id))
	}

	return fmt.Sprintf("%s\n%s", m.headerView(), m.viewport.View())
}

func (m detailPage) headerView() string {
	terminalWidth := max(m.viewport.Width, 1)
	w := min(max(terminalWidth-6, 24), 92)
	var p float64
	if m.contentDetail.Id > 0 {
		p = m.viewport.ScrollPercent() * 100
	}
	labelText := fmt.Sprintf("%3.f%%", p)
	label := styles.Meta.Render(labelText)
	crumb := m.contentDetail.Node.Title
	if m.contentDetail.Title != "" {
		crumb += "  /  " + m.contentDetail.Title
	}
	crumb = ansi.TruncateWc(crumb, max(w-lipgloss.Width(labelText)-3, 6), "…")
	gap := max(w-lipgloss.Width(crumb)-lipgloss.Width(labelText)-2, 1)
	top := styles.Title.Render(crumb) + strings.Repeat(" ", gap) + label
	trackWidth := w
	filled := min(trackWidth, max(0, int(math.Round(float64(trackWidth)*p/100))))
	track := styles.Active.Render(strings.Repeat("━", filled)) + styles.Divider.Render(strings.Repeat("─", trackWidth-filled))
	header := top + "\n" + track
	return lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, header)
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
		terminalWidth, _ = g.Window.GetSize()
		w                = min(max(terminalWidth-6, 24), 92)
		content          strings.Builder
		me               = g.Me.Get()
		decodeItems      []string
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

	// Topic header.
	content.WriteString(styles.Active.Bold(true).Render(m.contentDetail.Node.Title))
	content.WriteString("\n\n")
	content.WriteString(styles.Title.Width(w).Render(m.contentDetail.Title))
	content.WriteString("\n")
	meta := fmt.Sprintf("%s  ·  %s  ·  %d 回复", m.contentDetail.Member.GetUserNameLabel(me.Id), relativeTime(m.contentDetail.Created), m.contentDetail.Replies)
	meta = ansi.TruncateWc(meta, w, "…")
	content.WriteString(styles.Meta.Render(meta))
	content.WriteString("\n\n")
	content.WriteString(styles.Divider.Render(strings.Repeat("─", w)))
	content.WriteString("\n\n")
	content.WriteString(strings.TrimSpace(decodeFn(m.contentDetail.GetContent(w))))
	content.WriteString("\n\n")

	// Supplements are visually separated from the original post.
	for i, c := range m.contentDetail.Supplements {
		heading := styles.Active.Bold(true).Render(fmt.Sprintf("附言 #%d", i+1)) + styles.Meta.Render("  ·  "+relativeTime(c.Created))
		body := heading + "\n" + strings.TrimSpace(decodeFn(c.GetContent(max(w-3, 12))))
		box := lipgloss.NewStyle().
			Border(lipgloss.Border{Left: "┃"}, false, false, false, true).
			BorderForeground(styles.Theme.Accent).
			Background(styles.Theme.Surface).
			Padding(0, 1).
			Width(max(w-3, 12)).
			Render(body)
		content.WriteString(box)
		content.WriteString("\n\n")
	}

	content.WriteString(styles.Divider.Render(strings.Repeat("─", w)))
	content.WriteString("\n\n")
	loadedText := ""
	if len(m.contentReply) > 0 {
		loadedText = fmt.Sprintf("  已加载 %d", len(m.contentReply))
	}
	content.WriteString(styles.Title.Render(fmt.Sprintf("%d 条回复", m.contentDetail.Replies)) + styles.Meta.Render(loadedText))
	content.WriteString("\n\n")

	// Replies use a light separator instead of nested boxes.
	var replyContent strings.Builder
	for i, r := range m.contentReply {
		opText := lo.If(r.Member.Id == m.contentDetail.Member.Id, " "+styles.MemberOp).Else("")
		left := styles.Active.Bold(true).Render(fmt.Sprintf("#%d", i+1)) + "  " + styles.Title.Render(r.Member.GetUserNameLabel(me.Id)) + opText
		timeText := relativeTime(r.Created)
		left = ansi.TruncateWc(left, max(w-lipgloss.Width(timeText)-2, 8), "…")
		gap := max(w-lipgloss.Width(left)-lipgloss.Width(timeText), 1)
		replyContent.WriteString(left + strings.Repeat(" ", gap) + styles.Meta.Render(timeText))
		replyContent.WriteString("\n")
		replyContent.WriteString(strings.TrimSpace(decodeFn(r.GetContent(w))))
		if i < len(m.contentReply)-1 {
			replyContent.WriteString("\n\n")
			replyContent.WriteString(styles.Divider.Render(strings.Repeat("─", w)))
			replyContent.WriteString("\n\n")
		}
	}
	content.WriteString(replyContent.String())
	content.WriteString("\n\n")
	content.WriteString(m.replyStatus(w))

	column := lipgloss.NewStyle().Width(w).Render(content.String())
	return lipgloss.PlaceHorizontal(max(terminalWidth, 1), lipgloss.Center, column)
}

func (m detailPage) replyStatus(width int) string {
	var text string
	switch {
	case m.contentDetail.Replies == 0:
		text = "暂无回复"
	case m.replyPageInfo.CurrPage == 0:
		text = "正在加载评论…"
	case m.replyPageInfo.CurrPage < m.replyPageInfo.TotalPages:
		text = fmt.Sprintf("Enter  加载下一页评论   ·   %d / %d", len(m.contentReply), m.contentDetail.Replies)
	default:
		text = fmt.Sprintf("已加载全部评论   ·   %d / %d", len(m.contentReply), m.contentDetail.Replies)
	}
	return lipgloss.NewStyle().
		Foreground(styles.Theme.Muted).
		Background(styles.Theme.Surface).
		Align(lipgloss.Center).
		Width(width).
		Padding(0, 1).
		Render(text)
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
					width      = min(max(w-6, 24), 92)
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
