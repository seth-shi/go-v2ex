package pages

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/api"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/nav"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/seth-shi/go-v2ex/v2/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	inactiveTabStyle              = styles.Meta.Padding(0, 1)
	activeTabStyle                = lipgloss.NewStyle().Foreground(styles.Theme.Text).Background(styles.Theme.AccentDim).Bold(true).Padding(0, 1)
	_                nav.PageLife = topicPage{}
)

type topicPage struct {
	topics     []response.TopicResult
	index      int
	page       int
	totalPages int
	loading    bool
	cachePages int
	firstText  string
}

func newTopicPage() topicPage {
	return topicPage{
		page:       1,
		totalPages: 1,
		loading:    true,
	}
}

func (m topicPage) Init() tea.Cmd {
	// 获取第一页的数据
	return m.getTopics(m.page)
}

func (m topicPage) OnEntering() (tea.Model, tea.Cmd) {
	return m, commands.Post(messages.ShowStatusBarTextRequest{FirstText: m.firstText})
}

func (m topicPage) OnLeaving() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m topicPage) getTopics(page int) tea.Cmd {
	return commands.LoadingRequestTopics.Run(api.V2ex.GetTopics(context.Background(), page))
}

func (m topicPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case messages.GetTopicResponse:
		return m.onTopicResult(msg)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.Tab):
			return m.moveTabs(1)
		case key.Matches(msg, consts.AppKeyMap.ShiftTab):
			return m.moveTabs(-1)
		case key.Matches(msg, consts.AppKeyMap.KeyE):
			if len(m.topics) == 0 {
				return m, nil
			}
			// 查看详情
			curr := lo.NthOrEmpty(m.topics, m.index)
			// 去详情页面
			return m, nav.Push(newDetailPage(curr.Id))
		case key.Matches(msg, consts.AppKeyMap.Up):
			m.index--
			if m.index < 0 {
				m.index = len(m.topics) - 1
			}
			return m, nil
		case key.Matches(msg, consts.AppKeyMap.Down):
			m.index++
			if m.index > len(m.topics)-1 {
				m.index = 0
			}
			return m, nil
		case key.Matches(msg, consts.AppKeyMap.Left):
			if m.page > 1 {
				return m, m.getTopics(m.page - 1)
			}
			return m, nil
		case key.Matches(msg, consts.AppKeyMap.Right):
			if m.page < m.totalPages {
				return m, m.getTopics(m.page + 1)
			}
		default:
			return m, nil
		}
	}

	return m, nil
}

func (m topicPage) View() string {
	var doc strings.Builder
	header := m.renderTabs()
	doc.WriteString(header)
	doc.WriteString("\n")
	if m.loading {
		w, h := g.Window.GetSize()
		availableHeight := max(h-lipgloss.Height(header)-1, 1)
		doc.WriteString(loadingViewWithin("正在获取主题", w, availableHeight))
	} else {
		doc.WriteString(m.renderTables())
	}

	return doc.String()
}

func (m topicPage) moveTabs(add int) (tea.Model, tea.Cmd) {

	saveTabFn := func() tea.Msg {
		return g.Config.Save(
			func(conf *model.FileConfig) {
				conf.ActiveTab = g.TabNodeIndex(conf.ActiveTab, add)
			},
		)
	}

	m.page = 1
	return m, tea.Sequence(
		saveTabFn,
		m.getTopics(1),
	)
}

func (m topicPage) onTopicResult(msg messages.GetTopicResponse) (tea.Model, tea.Cmd) {

	var (
		result   = msg.Data
		pageInfo = msg.PageInfo
		pageText = fmt.Sprintf("第 %d 页 · %d 个主题", msg.PageInfo.CurrPage, msg.PageInfo.TotalCount)
	)

	m.cachePages = msg.CachePages
	m.topics = result
	// 会话的直接设置
	m.page = pageInfo.CurrPage
	m.totalPages = pageInfo.TotalPage()
	// 显示错误和页码
	m.loading = false

	m.firstText = pageText

	return m, commands.Post(messages.ShowStatusBarTextRequest{FirstText: m.firstText})
}

func (m topicPage) renderTabs() string {
	var (
		doc            strings.Builder
		renderedTabs   []string
		tabs           = g.OfficialNodes
		activeTabIndex = g.Config.Get().ActiveTab
		width, _       = g.Window.GetSize()
		contentWidth   = max(width-2, 20)
	)

	// Draw a compact, borderless tab strip. On narrow screens it follows the
	// active tab instead of wrapping and pushing the list out of alignment.
	start := max(0, activeTabIndex-2)
	end := start
	used := 0
	for end < len(tabs) {
		next := lipgloss.Width(tabs[end].Name) + 2
		if end > start && used+next > max(contentWidth-4, 20) {
			break
		}
		used += next
		end++
	}
	if start > 0 {
		renderedTabs = append(renderedTabs, styles.Meta.Render("‹ "))
	}
	for i := start; i < end; i++ {
		t := g.GetGroupNode(i)
		style := inactiveTabStyle
		if i == activeTabIndex {
			style = activeTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(t.Name))
	}
	if end < len(tabs) {
		renderedTabs = append(renderedTabs, styles.Meta.Render(" ›"))
	}

	active := g.GetGroupNode(activeTabIndex)
	pageText := fmt.Sprintf("第 %d 页", m.page)
	heading := styles.Title.Render("主题") + styles.Meta.Render("  ·  "+active.Title())
	heading = ansi.TruncateWc(heading, max(contentWidth-lipgloss.Width(pageText)-2, 8), "…")
	gap := max(contentWidth-lipgloss.Width(heading)-lipgloss.Width(pageText), 1)
	doc.WriteString(" " + heading + strings.Repeat(" ", gap) + styles.Meta.Render(pageText))
	doc.WriteString("\n ")
	doc.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, renderedTabs...))
	doc.WriteString("\n ")
	doc.WriteString(styles.Divider.Render(strings.Repeat("─", contentWidth)))

	return doc.String()
}

func (m topicPage) renderTables() string {
	if len(m.topics) == 0 {
		w, h := g.Window.GetSize()
		empty := lipgloss.JoinVertical(lipgloss.Center,
			styles.Title.Render("这里还没有主题"),
			styles.Meta.Render("可用 Tab 切换节点，或用 ← → 查看其他页面"),
		)
		return lipgloss.Place(max(w, 1), max(h-4, 1), lipgloss.Center, lipgloss.Center, empty)
	}
	var (
		w, h = g.Window.GetSize()
		me   = g.Me.Get()
		doc  strings.Builder
	)

	contentWidth := max(w-2, 16)
	// Medium and narrow terminals use a calmer two-line layout. Wide screens
	// keep the table so author and activity metadata stay easy to compare.
	if w >= 92 {
		return lipgloss.NewStyle().MarginLeft(1).Render(m.renderTopicTable(contentWidth, me.Id))
	}

	visible := max((h-4)/2, 1)
	start := min(max(m.index-visible/2, 0), max(len(m.topics)-visible, 0))
	end := min(start+visible, len(m.topics))
	for i := start; i < end; i++ {
		topic := m.topics[i]
		selected := i == m.index
		marker := "  "
		if selected {
			marker = styles.Active.Bold(true).Render("┃ ")
		}
		title := ansi.TruncateWc(topic.GetTitle(), max(contentWidth-2, 8), "…")
		titleStyle := styles.Title
		if selected {
			titleStyle = titleStyle.Foreground(styles.Theme.Accent)
		}
		member := topicUser(topic, me.Id)
		meta := fmt.Sprintf("%s · %s · %s回复 · %s", topic.Node.Title, member, formatMetric(topic.Replies), relativeTime(topic.LastTouched))
		meta = ansi.TruncateWc(meta, max(contentWidth-2, 8), "…")
		row := marker + titleStyle.Render(title) + "\n  " + styles.Meta.Render(meta)
		if selected {
			row = styles.SelectedRow.Width(contentWidth).Render(row)
		}
		doc.WriteString(row)
		if i < end-1 {
			doc.WriteString("\n")
		}
	}
	return lipgloss.NewStyle().MarginLeft(1).Render(doc.String())
}

func (m topicPage) renderTopicTable(width, meID int) string {
	const (
		nodeWidth    = 8
		memberWidth  = 12
		repliesWidth = 6
		activeWidth  = 8
	)
	var doc strings.Builder
	titleWidth := max(width-nodeWidth-memberWidth-repliesWidth-activeWidth-10, 12)
	header := "  " + fixedCell("节点", nodeWidth) + "  " + fixedCell("标题", titleWidth) + "  " + fixedCell("用户", memberWidth) + "  " + rightCell("回复", repliesWidth) + "  " + rightCell("活跃", activeWidth)
	doc.WriteString(styles.Meta.Bold(true).Render(header))
	doc.WriteString("\n")
	doc.WriteString(styles.Divider.Render(strings.Repeat("─", max(width, 1))))
	doc.WriteString("\n")

	_, h := g.Window.GetSize()
	visible := max(h-6, 1)
	start := min(max(m.index-visible/2, 0), max(len(m.topics)-visible, 0))
	end := min(start+visible, len(m.topics))
	for i := start; i < end; i++ {
		topic := m.topics[i]
		selected := i == m.index
		marker := "  "
		if selected {
			marker = styles.Active.Bold(true).Render("┃ ")
		}
		title := fixedCell(topic.GetTitle(), titleWidth)
		if selected {
			title = styles.Active.Bold(true).Render(title)
		}
		row := marker +
			styles.Meta.Render(fixedCell(topic.Node.Title, nodeWidth)) + "  " +
			title + "  " +
			fixedCell(topicUser(topic, meID), memberWidth) + "  " +
			rightCell(formatMetric(topic.Replies), repliesWidth) + "  " +
			styles.Meta.Render(rightCell(relativeTime(topic.LastTouched), activeWidth))
		if selected {
			row = styles.SelectedRow.Width(max(width, 1)).Render(row)
		}
		doc.WriteString(row)
		if i < end-1 {
			doc.WriteString("\n")
		}
	}
	return doc.String()
}

func topicUser(topic response.TopicResult, meID int) string {
	if topic.Member.Username != "" {
		return topic.Member.GetUserNameLabel(meID)
	}
	if topic.LastReplyBy != "" {
		return "↩ " + topic.LastReplyBy
	}
	return "—"
}

func fixedCell(value string, width int) string {
	value = ansi.TruncateWc(value, width, "…")
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

func rightCell(value string, width int) string {
	value = ansi.TruncateWc(value, width, "…")
	return strings.Repeat(" ", max(width-lipgloss.Width(value), 0)) + value
}

func formatMetric(value int) string {
	switch {
	case value >= 1_000_000:
		return fmt.Sprintf("%.1fm", float64(value)/1_000_000)
	case value >= 1_000:
		return fmt.Sprintf("%.1fk", float64(value)/1_000)
	default:
		return fmt.Sprintf("%d", value)
	}
}
