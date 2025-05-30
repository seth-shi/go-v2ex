package home

import (
	"strconv"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dromara/carbon/v2"
	"github.com/seth-shi/go-v2ex/internal/ui/context"
	"github.com/seth-shi/go-v2ex/internal/ui/events"
)

var (
	cellStyle   = lipgloss.NewStyle().Padding(0, 1).Width(5).Foreground(lipgloss.Color("#c2c2c2"))
	tableStyles = map[int]lipgloss.Style{
		0: cellStyle.Width(4).Align(lipgloss.Left).MarginBottom(1),
		1: cellStyle.Width(10).Align(lipgloss.Left).MarginBottom(1),
		3: cellStyle.Width(20).Align(lipgloss.Left).MarginBottom(1),
		4: cellStyle.Width(22).Align(lipgloss.Left).MarginBottom(1),
		5: cellStyle.Width(3).Align(lipgloss.Center).MarginBottom(1),
	}
	headerStyle = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center)
)

type Model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	ctx        *context.Data
}

func New(ctx *context.Data) Model {
	return Model{
		ctx: ctx,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {

	if len(m.ctx.Topics) == 0 {
		return ""
	}

	var rows [][]string
	for i := 0; i < len(m.ctx.Topics); i++ {
		topic := m.ctx.Topics[0]
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
	titleWidth := m.ctx.ScreenWidth
	for _, v := range tableStyles {
		titleWidth -= v.GetWidth()
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
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

				if row == m.ctx.TopicIndex {
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if m.ctx.LoadingText != nil {
		return m, nil
	}

	switch typeMsg := msg.(type) {
	case tea.KeyMsg:
		switch typeMsg.Type {
		case tea.KeyUp:
			m.ctx.TopicIndex--
			if m.ctx.TopicIndex < 0 {
				m.ctx.TopicIndex = max(0, len(m.ctx.Topics)-1)
			}
			return m, nil
		case tea.KeyDown:

			m.ctx.TopicIndex++
			if m.ctx.TopicIndex > len(m.ctx.Topics) {
				m.ctx.TopicIndex = 0
			}
			return m, nil
		case tea.KeyLeft:
			if m.ctx.TopicPage > 0 {
				m.ctx.TopicPage--
				return m, events.GetTopics(m.ctx.TopicPage)
			}
			return m, nil

		case tea.KeyRight:
			m.ctx.TopicPage++
			return m, events.GetTopics(m.ctx.TopicPage)
		default:
			return m, nil
		}
	}

	return m, nil
}
