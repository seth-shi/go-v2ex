package topics

import (
	"errors"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/consts"

	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dromara/carbon/v2"
)

var (
	cellStyle   = lipgloss.NewStyle().Padding(0, 1).Width(5)
	tableStyles = map[int]lipgloss.Style{
		0: cellStyle.Width(4).Align(lipgloss.Left),
		1: cellStyle.Width(10).Align(lipgloss.Left),
		3: cellStyle.Width(20).Align(lipgloss.Left),
		4: cellStyle.Width(22).Align(lipgloss.Left),
		5: cellStyle.Width(7).Align(lipgloss.Center),
	}
	headerStyle       = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center)
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
)

type Model struct {
	activeIndex int
	activeTab   int
	topicsPage  int
	requesting  bool
	topics      []*types.TopicResource
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msgType := msg.(type) {
	// 其它地方负责回调这里去请求数据,
	case messages.GetTopicsRequest:
		m.topicsPage = msgType.Page
		m.requesting = true
		// 默认进来是要给节点
		m.activeTab = msgType.NodeIndex
		return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.GetTopics(msgType.NodeIndex, msgType.Page))
	case messages.GetTopicsResult:
		m.topics = msgType.Topics
		m.requesting = false
		return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.End), messages.Post(msgType.Error))
	case tea.KeyMsg:
		// 如果在请求中, 不处理键盘事件
		if m.requesting {
			return m, messages.Post(errors.New("请求中"))
		}

		if key.Matches(msgType, consts.AppKeyMap.Tab) {
			m.activeTab++
			if m.activeTab >= len(config.G.GetNodes()) {
				m.activeTab = 0
			}
			return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.GetTopics(m.activeTab, m.topicsPage))
		}

		switch msgType.Type {
		case tea.KeyUp:
			m.activeIndex--
			if m.activeIndex < 0 {
				m.activeIndex = max(0, len(m.topics)-1)
			}
			return m, nil
		case tea.KeyDown:
			m.activeIndex++
			if m.activeIndex >= len(m.topics) {
				m.activeIndex = 0
			}
			return m, nil
		case tea.KeyLeft:
			if m.topicsPage > 0 {
				m.topicsPage--
				m.requesting = true
				return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.GetTopics(m.activeTab, m.topicsPage))
			}
			return m, nil
		case tea.KeyRight:
			m.topicsPage++
			m.requesting = true
			return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.GetTopics(m.activeTab, m.topicsPage))
		default:
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {

	var (
		doc strings.Builder
	)
	doc.WriteString(m.renderTabs())
	doc.WriteString(m.renderTables())
	return doc.String()
}

func (m Model) renderTabs() string {
	var (
		doc          strings.Builder
		renderedTabs []string
		tabs         = config.G.GetNodes()
	)

	for i, t := range tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	return doc.String()
}

func (m Model) renderTables() string {
	if len(m.topics) == 0 {
		return ""
	}
	// 表格
	var rows [][]string
	for i := 0; i < len(m.topics); i++ {
		topic := m.topics[0]
		rows = append(
			rows, []string{
				strconv.Itoa(i + 1),
				topic.Node.Title,
				topic.Title,
				topic.Member.Username,
				carbon.CreateFromTimestamp(topic.LastTouched).String(),
				strconv.Itoa(topic.Replies),
			},
		)
	}

	// len(tableStyles) + 1 = 列数 (再 +1 等于边框数)
	titleWidth := config.Screen.Width - (len(tableStyles) + 1 + 1)
	for _, v := range tableStyles {
		titleWidth -= v.GetWidth()
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle()).
		StyleFunc(
			func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return headerStyle
				}

				style := cellStyle
				if s, e := tableStyles[col]; e {
					style = s
				} else if col == 2 {
					style = lipgloss.NewStyle().Width(titleWidth)
				}

				if row == m.activeIndex {
					style = style.Foreground(lipgloss.Color("#1e9fff")).Bold(true)
					rows[row][0] = "*"
				}

				return style
			},
		).
		Headers("#", "node", "title", "member", "last_touched", "replies").
		Rows(rows...)
	return t.String()
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}
