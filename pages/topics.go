package pages

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/x/ansi"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/seth-shi/go-v2ex/v2/styles"

	"github.com/seth-shi/go-v2ex/v2/api"
	"github.com/seth-shi/go-v2ex/v2/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dromara/carbon/v2"
)

const keyHelp = "[a/d:翻页 w/s:移动 e:详情 tab/shift+tab:节点 空格:老板键 ?:帮助页 `:设置页  =:显示页脚]"

var (
	cellStyle         = styles.Primary.Padding(0, 1).Width(5)
	headerStyle       = styles.Primary.Bold(true).Align(lipgloss.Center)
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	inactiveTabStyle  = styles.Primary.Border(inactiveTabBorder, true).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
)

type topicPage struct {
	topics     []response.TopicResult
	index      int
	page       int
	totalPages int
	loading    bool
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

func (m topicPage) getTopics(page int) tea.Cmd {
	return tea.Sequence(
		messages.LoadingRequestTopics.PostStart(),
		api.V2ex.GetTopics(context.Background(), page),
		messages.LoadingRequestTopics.PostEnd(),
	)
}

func (m topicPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msgType := msg.(type) {
	case messages.GetTopicResponse:
		return m.onTopicResult(msgType)
	case tea.KeyMsg:
		// 如果在请求中, 不处理键盘事件
		if messages.LoadingRequestTopics.Loading() {
			return m, commands.Post(errors.New("请求中"))
		}

		switch {
		case key.Matches(msgType, consts.AppKeyMap.Tab):
			return m.moveTabs(1)
		case key.Matches(msgType, consts.AppKeyMap.ShiftTab):
			return m.moveTabs(-1)
		case key.Matches(msgType, consts.AppKeyMap.KeyE):
			// 查看详情
			curr := lo.NthOrEmpty(m.topics, m.index)
			// 去详情页面
			return m, tea.Sequence(
				commands.Redirect(RouteDetail),
				commands.Post(messages.GetDetailRequest{ID: curr.Id}),
			)
		case key.Matches(msgType, consts.AppKeyMap.Up):
			m.index--
			if m.index < 0 {
				m.index = len(m.topics) - 1
			}
			return m, nil
		case key.Matches(msgType, consts.AppKeyMap.Down):
			m.index++
			if m.index > len(m.topics)-1 {
				m.index = 0
			}
			return m, nil
		case key.Matches(msgType, consts.AppKeyMap.Left):
			if m.page > 1 {
				return m, m.getTopics(m.page - 1)
			}
			return m, nil
		case key.Matches(msgType, consts.AppKeyMap.Right):
			if m.page < m.totalPages {
				return m, m.getTopics(m.page + 1)
			}
			// 如果用 V1, 提醒可以是用 R 键切换
			canChoose := g.GetGroupNode(g.Config.Get().ActiveTab).CanChooseApiVersion()
			if !g.Session.IsApiV2.Load() && canChoose {
				return m, commands.AlertInfo(
					fmt.Sprintf(
						"按[%s]键可切换接口版本", strings.Join(
							consts.AppKeyMap.KeyR.Keys(),
							" ",
						),
					),
				)
			}

		case key.Matches(msgType, consts.AppKeyMap.KeyR):
			// 切换 V2 接口
			g.Session.ChooseApiV2.Store(!g.Session.ChooseApiV2.Load())
			return m, m.getTopics(1)
		default:
			return m, nil
		}
	}

	return m, nil
}

func (m topicPage) View() string {

	if m.loading {
		return loadingView("获取列表数据中...")
	}

	var (
		doc strings.Builder
	)
	doc.WriteString(m.renderTabs())
	doc.WriteString("\n")
	doc.WriteString(m.renderTables())
	return doc.String()
}

func (m topicPage) moveTabs(add int) (tea.Model, tea.Cmd) {

	saveTabFn := func() tea.Msg {
		return g.Config.Save(
			func(conf *model.FileConfig) {
				slog.Info("save index")
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

func (m topicPage) onTopicResult(msgType messages.GetTopicResponse) (tea.Model, tea.Cmd) {

	var (
		result   = msgType.Data
		pageInfo = msgType.PageInfo
		apiText  = "v1@api"
	)

	m.topics = result
	// 会话的直接设置
	m.page = pageInfo.CurrPage
	m.totalPages = pageInfo.TotalPage()
	// 显示错误和页码
	m.loading = false

	if g.Session.IsApiV2.Load() {
		apiText = "v2@api"
	}

	return m, commands.Post(
		messages.ShowStatusBarTextRequest{
			FirstText: fmt.Sprintf("%s %s", apiText, pageInfo.ToString()),
			HelpText:  keyHelp,
		},
	)
}

func (m topicPage) renderTabs() string {
	var (
		doc            strings.Builder
		renderedTabs   []string
		tabs           = g.OfficialNodes
		activeTab      *g.GroupNode
		activeTabIndex = g.Config.Get().ActiveTab
	)

	for i, _ := range tabs {
		t := g.GetGroupNode(i)
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(tabs)-1, i == activeTabIndex
		if isActive {
			style = activeTabStyle
			activeTab = &t
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
		renderedTabs = append(renderedTabs, style.Render(t.Name))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	// 增加一行显示二级
	if activeTab != nil {
		doc.WriteString("\n")
		doc.WriteString(styles.Hint.PaddingLeft(1).Render(activeTab.Title()))
	}

	return doc.String()
}

func (m topicPage) renderTables() string {
	if len(m.topics) == 0 {
		return ""
	}
	// 表格
	var (
		w, _        = g.Window.GetSize()
		me          = g.Me.Get()
		headers     = []string{"#", "节点", "标题", "OP", "回复数", "时间"}
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
	if g.Session.IsApiV2.Load() && len(headers) > 3 {
		headers[3] = "LR"
	}

	for i, topic := range m.topics {

		// 设置列自适应宽度
		nodeTitle := topic.Node.Title
		if len(columnWidth) > 1 && len(nodeTitle) > columnWidth[1] {
			// lipgloss.Width 处理中文, len 处理空格
			columnWidth[1] = max(lipgloss.Width(nodeTitle), len(nodeTitle))
		}

		// 这样子就不会显示 OP
		member := topic.Member.GetUserNameLabel(me.Id)
		if len(columnWidth) > 3 && len(member) > columnWidth[3] {
			columnWidth[3] = min(lipgloss.Width(member), len(member)) + 3
		}

		rows = append(
			rows, []string{
				strconv.Itoa(i + 1),
				nodeTitle,
				topic.GetTitle(),
				member,
				styles.HotText(topic.Replies),
				carbon.CreateFromTimestamp(topic.LastTouched).String(),
			},
		)
	}

	if len(columnWidth) > 1 && columnWidth[1] < 4 {
		columnWidth[1] = 4
	}

	// len(tableStyles) + 1 = 列数 (再 +1 等于边框数)
	titleWidth := w - (len(columnWidth) + 1 + 1) - lo.Sum(columnWidth)
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

				if row == m.index {
					style = styles.Active.Bold(true)
					rows[row][0] = "*"
				}

				return style
			},
		).
		Headers(headers...).
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
