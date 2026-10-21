package pages

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/seth-shi/go-v2ex/v2/api"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/nav"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

var _ nav.PageLife = topicPage{}

type topicPage struct {
	topics                             []response.TopicResult
	index, page, totalPages, activeTab int
	node                               g.GroupNode
	loading                            bool
	loadErr                            error
	requestID                          uint64
	pendingPage                        int
	fresh                              bool
	refreshedAt                        time.Time
	firstText                          string
}

func newTopicPage() topicPage {
	tab := g.Config.Get().ActiveTab
	return topicPage{page: 1, pendingPage: 1, totalPages: 1, activeTab: tab, node: g.GetGroupNode(tab), loading: true, requestID: nextRequestID()}
}
func (m topicPage) Init() tea.Cmd { return tea.Batch(m.getTopics(1, false), loadHistory()) }
func (m topicPage) OnEntering() (tea.Model, tea.Cmd) {
	latest := g.GetGroupNode(m.activeTab)
	if !slices.Equal(m.node.Nodes, latest.Nodes) {
		m.node = latest
		m.topics, m.index, m.page, m.totalPages = nil, 0, 1, 1
		m.refreshedAt = time.Time{}
		return m.request(1, true)
	}
	return m, commands.Post(messages.ShowStatusBarTextRequest{Page: "topicPage", FirstText: m.firstText})
}
func (m topicPage) OnLeaving() (tea.Model, tea.Cmd) { return m, nil }
func (m topicPage) AcceptsMessage(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case messages.GetTopicResponse:
		return msg.RequestID == m.requestID && msg.NodeKey == m.node.Key
	case messages.RequestError:
		return msg.Kind == messages.RequestTopics && msg.RequestID == m.requestID && msg.NodeKey == m.node.Key
	}
	return false
}
func (m topicPage) getTopics(page int, fresh bool) tea.Cmd {
	return api.V2ex.GetTopicsForNode(context.Background(), m.node, page, m.requestID, fresh)
}
func (m topicPage) request(page int, fresh bool) (tea.Model, tea.Cmd) {
	m.requestID = nextRequestID()
	m.pendingPage = page
	m.loading, m.fresh, m.loadErr = true, fresh, nil
	return m, m.getTopics(page, fresh)
}
func (m topicPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.GetTopicResponse:
		if msg.RequestID != m.requestID || msg.NodeKey != m.node.Key {
			return m, nil
		}
		return m.onTopicResult(msg)
	case messages.RequestError:
		if msg.Kind != messages.RequestTopics || msg.RequestID != m.requestID || msg.NodeKey != m.node.Key {
			return m, nil
		}
		m.loading, m.loadErr = false, msg.Err
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.Refresh):
			if m.loading {
				return m, nil
			}
			page := m.page
			if m.loadErr != nil {
				page = m.pendingPage
			}
			return m.request(page, true)
		case key.Matches(msg, consts.AppKeyMap.Tab):
			return m.moveTabs(1)
		case key.Matches(msg, consts.AppKeyMap.ShiftTab):
			return m.moveTabs(-1)
		case key.Matches(msg, consts.AppKeyMap.KeyE):
			if len(m.topics) == 0 {
				return m, nil
			}
			return m, nav.Push(newDetailPage(m.topics[m.index].Id))
		case key.Matches(msg, consts.AppKeyMap.Up):
			if len(m.topics) > 0 {
				m.index = (m.index + len(m.topics) - 1) % len(m.topics)
			}
		case key.Matches(msg, consts.AppKeyMap.Down):
			if len(m.topics) > 0 {
				m.index = (m.index + 1) % len(m.topics)
			}
		case key.Matches(msg, consts.AppKeyMap.Left):
			if !m.loading && m.page > 1 {
				return m.request(m.page-1, false)
			}
		case key.Matches(msg, consts.AppKeyMap.Right):
			if !m.loading && m.page < m.totalPages {
				return m.request(m.page+1, false)
			}
		}
	}
	return m, nil
}
func (m topicPage) moveTabs(add int) (tea.Model, tea.Cmd) {
	m.activeTab = g.TabNodeIndex(m.activeTab, add)
	m.node = g.GetGroupNode(m.activeTab)
	m.topics, m.index, m.page, m.totalPages = nil, 0, 1, 1
	m.refreshedAt = time.Time{}
	tab := m.activeTab
	updated, request := m.request(1, false)
	return updated, tea.Batch(request, func() tea.Msg { return g.Config.Save(func(conf *model.FileConfig) { conf.ActiveTab = tab }) })
}
func (m topicPage) onTopicResult(msg messages.GetTopicResponse) (tea.Model, tea.Cmd) {
	if msg.PageInfo == nil {
		return m, nil
	}
	// A refresh can reduce the number of pages; follow the final valid page.
	if msg.PageInfo.CurrPage > max(msg.PageInfo.TotalPage(), 1) {
		if m.fresh {
			m.refreshedAt = time.Now()
		}
		return m.request(max(msg.PageInfo.TotalPage(), 1), false)
	}
	var selectedID int64
	if len(m.topics) > 0 && m.page == msg.PageInfo.CurrPage {
		selectedID = m.topics[m.index].Id
	}
	m.topics, m.index = msg.Data, 0
	for i, t := range m.topics {
		if t.Id == selectedID {
			m.index = i
			break
		}
	}
	m.page, m.totalPages = msg.PageInfo.CurrPage, max(msg.PageInfo.TotalPage(), 1)
	m.loading, m.loadErr = false, nil
	if m.fresh {
		m.refreshedAt = time.Now()
	}
	m.firstText = fmt.Sprintf("第 %d/%d 页 · %d 个主题", m.page, m.totalPages, msg.PageInfo.TotalCount)
	if !m.refreshedAt.IsZero() {
		m.firstText += " · 刷新于 " + m.refreshedAt.Format("15:04")
	}
	return m, commands.Post(messages.ShowStatusBarTextRequest{Page: "topicPage", FirstText: m.firstText})
}
func (m topicPage) View() string {
	w, h := pageSize()
	header := m.renderTabs()
	available := max(h-lipgloss.Height(header)-1, 1)
	var body string
	switch {
	case m.loading && len(m.topics) == 0:
		body = loadingViewWithin("正在获取主题", w, available)
	case m.loadErr != nil && len(m.topics) == 0:
		body = requestErrorView("主题加载失败", m.loadErr, w, available)
	default:
		body = m.renderTopics(available)
	}
	return fitPage(header+"\n"+body, w, h)
}
func (m topicPage) renderTabs() string {
	w, _ := pageSize()
	width, pad := listGeometry(w)
	status := fmt.Sprintf("第 %d / %d 页", m.page, m.totalPages)
	if m.loading {
		status = "正在加载…"
		if len(m.topics) > 0 {
			status = "正在刷新…"
		}
	}
	if m.loadErr != nil {
		status = "Ctrl+R 重试"
	}
	heading := joinEdges(styles.Title.Render("GO V2EX")+styles.Meta.Render(" / "+m.node.Name), styles.Meta.Render(status), width)
	var tabs []string
	start := max(0, m.activeTab-2)
	used := 0
	if start > 0 {
		tabs = append(tabs, styles.Meta.Render("‹ "))
		used = 2
	}
	for i := start; i < len(g.OfficialNodes); i++ {
		label := " " + g.OfficialNodes[i].Name + " "
		if used+lipgloss.Width(label) > width-2 {
			break
		}
		style := styles.Meta
		if i == m.activeTab {
			style = styles.SelectedRow
		}
		tabs = append(tabs, style.Render(label))
		used += lipgloss.Width(label)
		if i == len(g.OfficialNodes)-1 {
			used = -1
		}
	}
	if used >= 0 {
		tabs = append(tabs, styles.Meta.Render(" ›"))
	}
	return indentLines(heading+"\n"+ansi.Truncate(strings.Join(tabs, ""), width, "")+"\n"+styles.Divider.Render(strings.Repeat("─", width)), pad)
}
func (m topicPage) renderTopics(height int) string {
	w, _ := pageSize()
	width, pad := listGeometry(w)
	if len(m.topics) == 0 {
		return emptyState("这里还没有主题", "Tab 切换节点 · Ctrl+R 刷新", w, height)
	}
	var lines []string
	if m.loadErr != nil {
		lines = append(lines, styles.Err.Render(ansi.Truncate("刷新失败，保留当前列表 · Ctrl+R 重试", width, "…")), "")
	}
	compact := w >= compactListMinWidth
	rowHeight := 2
	if compact {
		rowHeight = 1
		lines = append(lines, compactHeader([]string{"节点", "标题", "作者", "回复", "活跃"}, topicColumnWidths(width), width))
	}
	used := 0
	if len(lines) > 0 {
		used = lipgloss.Height(strings.Join(lines, "\n"))
	}
	visible := max((height-used)/rowHeight, 1)
	start, end := visibleRange(m.index, len(m.topics), visible)
	me := g.Me.Get()
	for i := start; i < end; i++ {
		t := m.topics[i]
		read := g.History.Contains(t.Id)
		if compact {
			values := []string{t.Node.Title, t.GetTitle(), topicUser(t, me.Id), formatMetric(t.Replies), relativeTime(t.LastTouched)}
			lines = append(lines, compactRow(values, topicColumnWidths(width), width, i == m.index, read))
			continue
		}
		meta := fmt.Sprintf("%s · %s · %s", cleanLine(t.Node.Title), topicUser(t, me.Id), relativeTime(t.LastTouched))
		if read {
			meta += " · 已读"
		}
		lines = append(lines, renderListItem(cleanLine(t.GetTitle()), meta, formatMetric(t.Replies), width, i == m.index, read))
	}
	return indentLines(strings.Join(lines, "\n"), pad)
}
func topicUser(topic response.TopicResult, meID int) string {
	if topic.Member.Username != "" {
		return cleanLine(topic.Member.GetUserNameLabel(meID))
	}
	if topic.LastReplyBy != "" {
		return "↩ " + cleanLine(topic.LastReplyBy)
	}
	return "—"
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
