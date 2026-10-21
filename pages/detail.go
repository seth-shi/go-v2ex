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
	"github.com/seth-shi/go-v2ex/v2/model"
)

var (
	_ nav.PageLife = detailPage{}
)

type detailPage struct {
	id               int64
	viewport         viewport.Model
	decodeMap        map[string]string
	requestID        uint64
	replyRequestID   uint64
	loading          bool
	replyLoading     bool
	loadErr          error
	replyErr         error
	pendingReplyPage int
	replacingReplies bool

	replyPageInfo response.V2PageResponse
	contentDetail response.V2DetailResult
	contentReply  []response.V2ReplyResult
}

func (m detailPage) OnEntering() (tea.Model, tea.Cmd) {
	// 获取内容 + 第一页的评论
	var (
		w, h = pageSize()
	)
	m.viewport.Width = w
	m.viewport.Height = max(h-lipgloss.Height(m.headerView())-1, 1)
	// 重新修改键盘映射
	m.viewport.KeyMap.Up = consts.AppKeyMap.Up
	m.viewport.KeyMap.Down = consts.AppKeyMap.Down
	m.viewport.KeyMap.Left = consts.AppKeyMap.Left
	m.viewport.KeyMap.Right = consts.AppKeyMap.Right
	if m.contentDetail.Id > 0 {
		m.refreshContent()
	}
	return m, m.statusCommand()
}

func (m detailPage) OnLeaving() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m detailPage) AcceptsMessage(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case messages.GetDetailResponse:
		return msg.TopicID == m.id && msg.RequestID == m.requestID
	case messages.GetReplyResponse:
		return msg.TopicID == m.id && msg.RequestID == m.replyRequestID
	case messages.DecodeDetailContentResult:
		return msg.TopicID == m.id && msg.RequestID == m.requestID
	case messages.RequestError:
		return msg.TopicID == m.id && ((msg.Kind == messages.RequestDetail && msg.RequestID == m.requestID) || (msg.Kind == messages.RequestReplies && msg.RequestID == m.replyRequestID))
	}
	return false
}

func newDetailPage(id int64) detailPage {
	return detailPage{
		id:        id,
		decodeMap: make(map[string]string),
		viewport:  viewport.New(40, 40),
		requestID: nextRequestID(),
		loading:   true,
	}
}

func (m detailPage) Init() tea.Cmd {
	return api.V2ex.GetDetail(context.Background(), m.id, m.requestID)
}

func (m detailPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = max(msg.Width, 1)
		_, height := pageSize()
		m.viewport.Height = max(height-lipgloss.Height(m.headerView())-1, 1)
		if m.contentDetail.Id > 0 {
			m.refreshContent()
		}
	case messages.GetDetailResponse:
		if msg.TopicID != m.id || msg.RequestID != m.requestID {
			return m, nil
		}
		m.loading = false
		if msg.Data.Id != m.id {
			m.loadErr = errors.New("接口未返回主题内容")
			return m, nil
		}
		m.loadErr = nil
		m.contentDetail = msg.Data
		m.decodeMap = make(map[string]string)
		m.replyErr = nil
		if m.contentDetail.Replies > 0 {
			m.replacingReplies = true
			cmds = append(cmds, m.startReply(1))
		} else {
			m.contentReply = nil
			m.replyPageInfo = response.V2PageResponse{}
			m.replyLoading = false
		}
		m.refreshContent()
		detail := m.contentDetail
		cmds = append(cmds, func() tea.Msg {
			err := g.History.Record(model.HistoryEntry{TopicID: detail.Id, Title: cleanLine(detail.Title), Node: cleanLine(detail.Node.Title), Author: cleanLine(detail.Member.Username), URL: detail.Url})
			return historyUpdated{Err: err}
		}, m.statusCommand())
	case messages.GetReplyResponse:
		if msg.TopicID != m.id || msg.RequestID != m.replyRequestID {
			return m, nil
		}
		m.replyLoading, m.replyErr = false, nil
		if m.replacingReplies {
			m.contentReply = nil
			m.replacingReplies = false
		}
		m.replyPageInfo = msg.Data.Pagination
		m.contentReply = appendUniqueReplies(m.contentReply, msg.Data.Result)
		m.refreshContent()
		cmds = append(cmds, m.statusCommand())
	case messages.RequestError:
		if msg.TopicID != m.id {
			return m, nil
		}
		switch {
		case msg.Kind == messages.RequestDetail && msg.RequestID == m.requestID:
			m.loading, m.loadErr = false, msg.Err
		case msg.Kind == messages.RequestReplies && msg.RequestID == m.replyRequestID:
			m.replyLoading, m.replyErr = false, msg.Err
		default:
			return m, nil
		}
		if m.contentDetail.Id > 0 {
			m.refreshContent()
		}
	case messages.DecodeDetailContentResult:
		if msg.TopicID != m.id || msg.RequestID != m.requestID {
			return m, nil
		}
		m.decodeMap = lo.Assign(m.decodeMap, msg.Result)
		m.refreshContent()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.Refresh):
			if m.loading {
				return m, nil
			}
			if m.replyErr != nil && m.loadErr == nil {
				cmd := m.startReply(m.pendingReplyPage)
				m.refreshContent()
				return m, cmd
			}
			m.loading, m.loadErr = true, nil
			m.requestID = nextRequestID()
			m.replyRequestID = nextRequestID()
			m.replyLoading = false
			return m, api.V2ex.GetDetail(context.Background(), m.id, m.requestID)
		case key.Matches(msg, consts.AppKeyMap.KeyE):
			if m.contentDetail.Id == 0 || m.loading || m.replyLoading {
				return m, nil
			}
			if m.replyErr != nil {
				cmd := m.startReply(m.pendingReplyPage)
				m.refreshContent()
				return m, cmd
			}
			if m.contentDetail.Replies == 0 || (m.replyPageInfo.TotalPages > 0 && m.replyPageInfo.CurrPage >= m.replyPageInfo.TotalPages) {
				return m, commands.Post(errors.New("已无更多评论"))
			}
			if m.replyPageInfo.CurrPage == 0 {
				return m, commands.Post(errors.New("评论正在加载中"))
			}
			cmd := m.startReply(m.replyPageInfo.CurrPage + 1)
			m.refreshContent()
			return m, cmd
		case key.Matches(msg, consts.AppKeyMap.KeyR):
			if m.contentDetail.Id == 0 || m.loading {
				return m, nil
			}
			return m.decodeContent()
		case key.Matches(msg, consts.AppKeyMap.F1):
			if m.contentDetail.Url == "" {
				return m, nil
			}
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
	w, h := pageSize()
	if m.contentDetail.Id == 0 {
		if m.loadErr != nil {
			return requestErrorView("主题加载失败", m.loadErr, w, h)
		}
		return loadingViewWithin(fmt.Sprintf("正在加载主题 #%d", m.id), w, h)
	}
	return fitPage(fmt.Sprintf("%s\n%s", m.headerView(), m.viewport.View()), w, h)
}

func (m detailPage) headerView() string {
	terminalWidth := max(m.viewport.Width, 1)
	w := readingWidth(terminalWidth)
	var p float64
	if m.contentDetail.Id > 0 {
		p = m.viewport.ScrollPercent() * 100
	}
	labelText := fmt.Sprintf("%3.f%%", p)
	if m.loading {
		labelText = "刷新中"
	}
	if m.loadErr != nil {
		labelText = "Ctrl+R 重试"
	}
	label := styles.Meta.Render(labelText)
	crumb := cleanLine(m.contentDetail.Node.Title)
	if m.contentDetail.Title != "" {
		crumb += "  /  " + cleanLine(m.contentDetail.Title)
	}
	top := joinEdges(styles.Title.Render(crumb), label, w)
	trackWidth := w
	filled := min(trackWidth, max(0, int(math.Round(float64(trackWidth)*p/100))))
	track := styles.Active.Render(strings.Repeat("━", filled)) + styles.Divider.Render(strings.Repeat("─", trackWidth-filled))
	header := top + "\n" + track
	return lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, header)
}

func (m *detailPage) startReply(page int) tea.Cmd {
	m.pendingReplyPage = max(page, 1)
	m.replyLoading, m.replyErr = true, nil
	m.replyRequestID = nextRequestID()
	return api.V2ex.GetReply(context.Background(), m.id, m.pendingReplyPage, m.replyRequestID)
}
func (m *detailPage) refreshContent() {
	offset := m.viewport.YOffset
	m.viewport.SetContent(m.buildContent())
	m.viewport.SetYOffset(offset)
}
func (m detailPage) statusCommand() tea.Cmd {
	return commands.Post(messages.ShowStatusBarTextRequest{Page: "detailPage", FirstText: m.replyPageInfo.ToString("0 评论")})
}
func readingWidth(width int) int { return min(max(width-4, 1), 88) }
func appendUniqueReplies(existing, incoming []response.V2ReplyResult) []response.V2ReplyResult {
	seen := make(map[int]bool)
	for _, r := range existing {
		if r.Id > 0 {
			seen[r.Id] = true
		}
	}
	for _, r := range incoming {
		if r.Id > 0 && seen[r.Id] {
			continue
		}
		existing = append(existing, r)
		if r.Id > 0 {
			seen[r.Id] = true
		}
	}
	return existing
}
func (m detailPage) buildContent() string {
	var (
		terminalWidth, _ = g.Window.GetSize()
		w                = readingWidth(terminalWidth)
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
	case m.replyErr != nil:
		text = "评论加载失败 · Enter 或 Ctrl+R 重试"
	case m.replyLoading:
		text = fmt.Sprintf("正在加载评论… · 已保留 %d 条", len(m.contentReply))
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
		Width(max(width-2, 1)).
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
					width      = readingWidth(w)
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
					TopicID:   m.id,
					RequestID: m.requestID,
					Result:    replaceMap,
				}
			},
		)
	return m, cmd
}
