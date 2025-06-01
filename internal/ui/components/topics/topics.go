package topics

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"

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
	cellStyle         = lipgloss.NewStyle().Padding(0, 1).Width(5)
	headerStyle       = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center)
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
)

type Model struct {
	activeIndex int
	activeTab   int
	page        int
	requesting  bool
	topics      []types.TopicComResult
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
		m.page = msgType.Page
		m.requesting = true
		// 默认进来是要给节点
		m.activeTab = msgType.NodeIndex
		return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.Client.GetTopics(msgType.NodeIndex, msgType.Page))
	case messages.GetTopicsResult:
		m.topics = msgType.Topics
		m.requesting = false
		return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.End), messages.Post(msgType.Error), messages.Post(messages.Tips{Text: fmt.Sprintf("第 %d 页", msgType.Page)}))
	case tea.KeyMsg:
		// 如果在请求中, 不处理键盘事件
		if m.requesting {
			return m, messages.Post(errors.New("请求中"))
		}

		if key.Matches(msgType, consts.AppKeyMap.Tab) {
			m.activeTab++
			m.page = 1
			if m.activeTab >= len(config.G.GetNodes()) {
				m.activeTab = 0
			}
			return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.Client.GetTopics(m.activeTab, m.page))
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
			if m.page > 0 {
				m.page--
				m.requesting = true
				return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.Client.GetTopics(m.activeTab, m.page))
			}
			return m, nil
		case tea.KeyRight:
			m.page++
			m.requesting = true
			return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.Client.GetTopics(m.activeTab, m.page))
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
	var (
		rows        [][]string
		columnWidth = []int{
			3, // 序号
			0,
			0,
			0,
			20, // 时间
			7,  // 回复数
		}
	)
	for i, topic := range m.topics {

		// 设置列自适应宽度
		if len(topic.Node) > columnWidth[1] {
			// lipgloss.Width 处理中文, len 处理空格
			columnWidth[1] = max(lipgloss.Width(topic.Node), len(topic.Node))
		}
		if len(topic.Member) > columnWidth[3] {
			columnWidth[3] = max(lipgloss.Width(topic.Member), len(topic.Member))
		}

		rows = append(
			rows, []string{
				strconv.Itoa(i + 1),
				topic.Node,
				topic.Title,
				topic.Member,
				carbon.CreateFromTimestamp(topic.LastTouched).String(),
				strconv.Itoa(topic.Replies),
			},
		)
	}

	// len(tableStyles) + 1 = 列数 (再 +1 等于边框数)
	titleWidth := config.Screen.Width - (len(columnWidth) + 1 + 1) - lo.Sum(columnWidth)
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle()).
		StyleFunc(
			func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return headerStyle
				}

				style := cellStyle
				if col == 2 {
					style = lipgloss.NewStyle().Width(titleWidth)
				} else if col < len(columnWidth)-1 {
					style = lipgloss.NewStyle().Width(columnWidth[col])
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
