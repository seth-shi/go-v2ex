package topics

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/model/response"

	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/model/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dromara/carbon/v2"
)

const keyHelp = "[a/d:翻页 w/s:选择 e:详情 tab/shift+tab:节点 空格:老板键 `:设置页 ?:帮助页 -:显示页脚]"

var (
	cellStyle         = lipgloss.NewStyle().Padding(0, 1).Width(5)
	headerStyle       = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center)
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
)

type Model struct {
	topics []response.TopicResult
}

func New() Model {
	return Model{}
}

func (m *Model) SetTopics(msg messages.GetTopicResponse) {
	m.topics = msg.Data.Items
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msgType := msg.(type) {
	// 其它地方负责回调这里去请求数据,
	case messages.GetTopicsRequest:
		// 默认进来是要给节点
		return m, tea.Sequence(
			messages.LoadingRequestTopics.PostStart(),
			api.V2ex.GetTopics(context.Background(), config.G.ActiveTab, msgType.Page),
			messages.LoadingRequestTopics.PostEnd(),
		)
	case messages.GetTopicResponse:
		return m, m.onTopicResult(msgType)
	case tea.KeyMsg:
		// 如果在请求中, 不处理键盘事件
		if messages.LoadingRequestTopics.Loading() {
			return m, messages.Post(errors.New("请求中"))
		}

		switch {
		case key.Matches(msgType, consts.AppKeyMap.Tab):
			return m, m.moveTabs(1)
		case key.Matches(msgType, consts.AppKeyMap.ShiftTab):
			return m, m.moveTabs(-1)
		case key.Matches(msgType, consts.AppKeyMap.Enter):
			// 查看详情
			curr := lo.NthOrEmpty(m.topics, config.Session.TopicActiveIndex)
			if curr.Id == 0 {
				return m, messages.Post(errors.New("查看无效的主题"))
			}
			return m, messages.Post(messages.RedirectDetailRequest{Id: curr.Id})
		case key.Matches(msgType, consts.AppKeyMap.Up):
			config.Session.TopicActiveIndex--
			if config.Session.TopicActiveIndex < 0 {
				config.Session.TopicActiveIndex = max(0, len(m.topics)-1)
			}
			return m, nil
		case key.Matches(msgType, consts.AppKeyMap.Down):
			config.Session.TopicActiveIndex++
			if config.Session.TopicActiveIndex >= len(m.topics) {
				config.Session.TopicActiveIndex = 0
			}
			return m, nil
		case key.Matches(msgType, consts.AppKeyMap.Left):
			if config.Session.TopicPage > 1 {
				return m, messages.Post(messages.GetTopicsRequest{Page: config.Session.TopicPage - 1})
			}
			return m, nil
		case key.Matches(msgType, consts.AppKeyMap.Right):
			return m, messages.Post(messages.GetTopicsRequest{Page: config.Session.TopicPage + 1})
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

func (m *Model) moveTabs(add int) tea.Cmd {
	config.Session.TopicPage = 1
	config.G.ActiveTab += add

	nodesSize := len(config.G.GetNodes())
	if nodesSize == 0 {
		return nil
	}

	if config.G.ActiveTab >= nodesSize {
		config.G.ActiveTab = 0
	}
	if config.G.ActiveTab < 0 {
		config.G.ActiveTab = nodesSize - 1
	}

	return tea.Batch(
		config.SaveToFile(""),
		messages.Post(messages.GetTopicsRequest{Page: config.Session.TopicPage}),
	)
}

func (m *Model) onTopicResult(msgType messages.GetTopicResponse) tea.Cmd {
	m.topics = msgType.Data.Items
	config.Session.TopicPage = msgType.Data.Pagination.CurrPage
	// 显示错误和页码
	pageInfo := msgType.Data.Pagination.ToString()
	return messages.Post(
		messages.ShowAlertRequest{
			Text: pageInfo,
			Help: keyHelp,
		},
	)
}

func (m *Model) renderTabs() string {
	var (
		doc          strings.Builder
		renderedTabs []string
		tabs         = config.G.GetNodes()
	)

	for i, t := range tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(tabs)-1, i == config.G.ActiveTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(lo.ValueOr(config.NodeMap, t, t)))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	return doc.String()
}

func (m *Model) renderTables() string {
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
			7,  // 回复数
			20, // 时间
		}
	)
	for i, topic := range m.topics {

		// 设置列自适应宽度
		nodeTitle := lo.ValueOr(config.NodeMap, topic.Node, topic.Node)
		if len(columnWidth) > 1 && len(nodeTitle) > columnWidth[1] {
			// lipgloss.Width 处理中文, len 处理空格
			columnWidth[1] = max(lipgloss.Width(nodeTitle), len(nodeTitle))
		}
		if len(columnWidth) > 3 && len(topic.Member) > columnWidth[3] {
			columnWidth[3] = max(lipgloss.Width(topic.Member), len(topic.Member))
		}

		rows = append(
			rows, []string{
				strconv.Itoa(i + 1),
				nodeTitle,
				topic.Title,
				topic.Member,
				strconv.Itoa(topic.Replies),
				carbon.CreateFromTimestamp(topic.LastTouched).String(),
			},
		)
	}

	// len(tableStyles) + 1 = 列数 (再 +1 等于边框数)
	titleWidth := config.Screen.Width - (len(columnWidth) + 1 + 1) - lo.Sum(columnWidth)
	for i, lines := range rows {
		if len(lines) > 2 {
			title := lines[2]
			rows[i][2] = ansi.TruncateWc(title, titleWidth, "...")
		}
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
				if col == 2 {
					style = lipgloss.NewStyle().Width(titleWidth)
				} else if col < len(columnWidth) {
					style = lipgloss.NewStyle().Width(columnWidth[col])
				}

				if row == config.Session.TopicActiveIndex {
					style = style.Foreground(lipgloss.Color("#1e9fff")).Bold(true)
					rows[row][0] = "*"
				}

				return style
			},
		).
		Headers("#", "节点", "标题", "member", "评论数", "最后回复时间").
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
