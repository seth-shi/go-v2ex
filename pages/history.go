package pages

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/nav"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

type historyPage struct {
	entries []model.HistoryEntry
	index   int
	loaded  bool
	loadErr error
}

func newHistoryPage() historyPage   { return historyPage{} }
func (m historyPage) Init() tea.Cmd { return nil }
func (m historyPage) OnEntering() (tea.Model, tea.Cmd) {
	var selectedID int64
	if len(m.entries) > 0 {
		selectedID = m.entries[m.index].TopicID
	}
	m.entries = g.History.Entries()
	m.index = min(m.index, max(len(m.entries)-1, 0))
	for i, entry := range m.entries {
		if entry.TopicID == selectedID {
			m.index = i
			break
		}
	}
	return m, tea.Batch(loadHistory(), commands.Post(messages.ShowStatusBarTextRequest{Page: "historyPage", FirstText: "本地历史"}))
}
func (m historyPage) OnLeaving() (tea.Model, tea.Cmd) { return m, nil }
func (m historyPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case historyUpdated:
		m.loaded, m.loadErr = true, msg.Err
		var selectedID int64
		if len(m.entries) > 0 {
			selectedID = m.entries[m.index].TopicID
		}
		m.entries, m.index = g.History.Entries(), 0
		for i, e := range m.entries {
			if e.TopicID == selectedID {
				m.index = i
				break
			}
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.Refresh):
			return m, loadHistory()
		case key.Matches(msg, consts.AppKeyMap.Up):
			if len(m.entries) > 0 {
				m.index = (m.index + len(m.entries) - 1) % len(m.entries)
			}
		case key.Matches(msg, consts.AppKeyMap.Down):
			if len(m.entries) > 0 {
				m.index = (m.index + 1) % len(m.entries)
			}
		case key.Matches(msg, consts.AppKeyMap.KeyE):
			if len(m.entries) > 0 {
				return m, nav.Push(newDetailPage(m.entries[m.index].TopicID))
			}
		}
	}
	return m, nil
}
func (m historyPage) View() string {
	w, h := pageSize()
	width, pad := listGeometry(w)
	heading := joinEdges(styles.Title.Render("阅读历史"), styles.Meta.Render(fmt.Sprintf("%d / 200", len(m.entries))), width)
	header := indentLines(heading+"\n"+styles.Meta.Render("仅保存在本机 · 最近打开的主题")+"\n"+styles.Divider.Render(strings.Repeat("─", width)), pad)
	available := max(h-lipgloss.Height(header)-1, 1)
	var body string
	switch {
	case !m.loaded && len(m.entries) == 0:
		body = loadingViewWithin("正在读取本地历史", w, available)
	case m.loadErr != nil:
		body = requestErrorView("历史文件读取失败，原文件已保留", m.loadErr, w, available)
	case len(m.entries) == 0:
		body = emptyState("还没有阅读历史", "打开主题后会自动记录 · Q 返回", w, available)
	default:
		var lines []string
		compact := w >= compactListMinWidth
		rowHeight := 2
		if compact {
			rowHeight = 1
			lines = append(lines, compactHeader([]string{"节点", "标题", "作者", "最近打开"}, historyColumnWidths(width), width))
		}
		remaining := available
		if compact {
			remaining -= 2
		}
		visible := max(remaining/rowHeight, 1)
		start, end := visibleRange(m.index, len(m.entries), visible)
		for i := start; i < end; i++ {
			e := m.entries[i]
			if compact {
				values := []string{e.Node, e.Title, e.Author, relativeTime(e.VisitedAt)}
				lines = append(lines, compactRow(values, historyColumnWidths(width), width, i == m.index, false))
				continue
			}
			meta := fmt.Sprintf("%s · %s · %s打开", cleanLine(e.Node), cleanLine(e.Author), relativeTime(e.VisitedAt))
			lines = append(lines, renderListItem(cleanLine(e.Title), meta, "", width, i == m.index, false))
		}
		body = indentLines(strings.Join(lines, "\n"), pad)
	}
	return fitPage(header+"\n"+body, w, h)
}
