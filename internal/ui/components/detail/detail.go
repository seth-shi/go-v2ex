package detail

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type Model struct {
	viewport viewport.Model
	ID       int
}

func New() Model {
	return Model{viewport: viewport.New(config.Screen.Width-2, config.Screen.Height-4)}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg.(type) {
	case messages.GetDetailRequest:
		m.viewport.Width = config.Screen.Width - 2
		m.viewport.Height = config.Screen.Height - 4
		var str strings.Builder
		for i := 0; i < 1000; i++ {
			str.WriteString(strings.Repeat(strconv.Itoa(i), 10))
		}
		m.viewport.SetContent(wrap.String(str.String(), config.Screen.Width))
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	m.viewport.Width = config.Screen.Width - 2
	m.viewport.Height = config.Screen.Height - 4
	m.viewport.YPosition = 2

	return fmt.Sprintf("%s\n%s", m.headerView(), m.viewport.View())
}

func (m Model) headerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
