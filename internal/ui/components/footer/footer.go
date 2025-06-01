package footer

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/seth-shi/go-v2ex/internal/config"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

const (
	rightText = "go-v2ex@v1.0.0 Powered by seth-shi"
)

type Model struct {
	focusIndex int
	// 数据
	// 只在 update view 读写, 无需上锁
	loadings map[int]string
	errors   []string
	leftText string
	spinner  spinner.Model
}

func New() Model {

	return Model{
		// 最大加载数限定
		loadings: make(map[int]string, 10),
		spinner:  spinner.New(spinner.WithSpinner(spinner.Points)),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msgType := msg.(type) {
	case messages.StartLoading:
		m.loadings[msgType.ID] = msgType.Text
		return m, nil
	case messages.EndLoading:
		delete(m.loadings, msgType.ID)
		return m, nil
	case messages.Tips:
		m.leftText = msgType.Text
		return m, nil
	case error:
		return m, m.addError(msgType)
	case messages.ClearErrorRequest:
		// 删除第一个元素
		m.errors = lo.Slice(m.errors, 1, len(m.errors))
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msgType)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {

	var (
		leftSection  []string
		rightSection = lipgloss.NewStyle().SetString(rightText)
	)

	if len(m.errors) > 0 || len(m.loadings) > 0 || m.leftText != "" {

		if m.leftText != "" {
			leftSection = append(leftSection, m.leftText)
		}

		leftSection = append(leftSection, lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff5722")).
			Render(strings.Join(m.errors, " / ")))

		loadingKeys := lo.Keys(m.loadings)
		slices.Sort(loadingKeys)
		loadingText := lo.Map(loadingKeys, func(key int, index int) string {
			return fmt.Sprintf(
				"%s %s",
				lipgloss.NewStyle().PaddingLeft(1).Render(
					m.spinner.View(),
				),
				m.loadings[key],
			)
		})
		leftSection = append(leftSection, lipgloss.NewStyle().Render(strings.Join(loadingText, "")))
	} else {
		helpKey := consts.AppKeyMap.HelpPage.Help()
		leftSection = append(leftSection, fmt.Sprintf("%s %s", helpKey.Key, helpKey.Desc))
	}

	padding := 1
	leftContent := strings.Join(leftSection, " ")
	footer := leftContent
	if config.G.ShowFooter {
		footer = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftContent,
			lipgloss.PlaceHorizontal(
				config.Screen.Width-lipgloss.Width(leftContent)-2*padding,
				lipgloss.Right,
				rightSection.Render(),
			),
		)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Width(config.Screen.Width).
		PaddingLeft(padding).
		PaddingRight(padding).
		Render(footer)
}

func (m *Model) addError(err error) tea.Cmd {

	if err == nil {
		return nil
	}

	m.errors = append(m.errors, err.Error())
	// 3s 后删除一个
	return tea.Tick(time.Second*3, func(time.Time) tea.Msg {
		return messages.ClearErrorRequest{}
	})
}
