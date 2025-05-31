package topics

import (
	"strconv"

	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dromara/carbon/v2"
)

var (
	cellStyle   = lipgloss.NewStyle().Padding(0, 1).Width(5).Foreground(lipgloss.Color("#c2c2c2"))
	tableStyles = map[int]lipgloss.Style{
		0: cellStyle.Width(4).Align(lipgloss.Left),
		1: cellStyle.Width(10).Align(lipgloss.Left),
		3: cellStyle.Width(20).Align(lipgloss.Left),
		4: cellStyle.Width(22).Align(lipgloss.Left),
		5: cellStyle.Width(3).Align(lipgloss.Center),
	}
	headerStyle = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center)
)

type Model struct {
	selectedIndex int
	topicsPage    int
	requesting    bool
	topics        []*types.TopicResource
	screenHeight  int
	screenWidth   int
}

func New() Model {
	return Model{
		selectedIndex: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch typeMsg := msg.(type) {
	// 其它地方负责回调这里去请求数据,
	case messages.GetTopicsRequest:
		m.topicsPage = typeMsg.Page
		m.requesting = true
		return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.GetTopics(typeMsg.Page))
	case messages.GetTopicsResult:
		m.topics = typeMsg.Topics
		m.requesting = false
		return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.End), messages.Post(typeMsg.Error))
	case tea.WindowSizeMsg:
		m.screenHeight = typeMsg.Height
		m.screenWidth = typeMsg.Width
		return m, nil
	case tea.KeyMsg:
		// 如果在请求中, 不处理键盘事件
		if m.requesting {
			return m, nil
		}

		switch typeMsg.Type {
		case tea.KeyUp:
			m.selectedIndex--
			if m.selectedIndex < 0 {
				m.selectedIndex = max(0, len(m.topics)-1)
			}
			return m, nil
		case tea.KeyDown:
			m.selectedIndex++
			if m.selectedIndex > len(m.topics) {
				m.selectedIndex = 0
			}
			return m, nil
		case tea.KeyLeft:
			if m.topicsPage > 0 {
				m.topicsPage--
				m.requesting = true
				return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.GetTopics(m.topicsPage))
			}
			return m, nil
		case tea.KeyRight:
			m.topicsPage++
			m.requesting = true
			return m, tea.Batch(messages.Post(messages.LoadingRequestTopics.Start), api.GetTopics(m.topicsPage))
		default:
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {

	if len(m.topics) == 0 {
		return ""
	}

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
	titleWidth := m.screenWidth
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

				if row == m.selectedIndex {
					style = style.Foreground(lipgloss.Color("#000000")).Bold(true)
					rows[row][0] = "*"
				}

				return style
			},
		).
		Headers("#", "node", "title", "member", "created", "last_touched", "replies").
		Rows(rows...)
	return t.Render()
}
