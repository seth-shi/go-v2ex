package footer

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

const (
	rightText = "go-v2ex Powered by seth-shi"
)

type Model struct {
	focusIndex int
	// 数据
	me         *types.V2MemberResult
	topicsPage int
	// 只在 update view 读写, 无需上锁
	loadings map[int]string
	errors   []string
	spinner  spinner.Model
	screen   types.ScreenSize
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

	switch typeMsg := msg.(type) {
	case messages.StartLoading:
		m.loadings[typeMsg.ID] = typeMsg.Text
		return m, nil
	case messages.EndLoading:
		delete(m.loadings, typeMsg.ID)
		return m, nil
	case error:
		return m, m.addError(typeMsg)
	case messages.ClearErrorRequest:
		// 删除第一个元素
		m.errors = lo.Slice(m.errors, 1, len(m.errors))
		return m, nil
	case messages.GetTopicsResult:
		m.topicsPage = typeMsg.Page
		return m, nil
	case tea.WindowSizeMsg:
		m.screen.Sync(typeMsg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(typeMsg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {

	var (
		leftSection  = lipgloss.NewStyle()
		rightSection = lipgloss.NewStyle().SetString(rightText)
	)

	// 如果有错误
	if len(m.errors) > 0 {
		leftSection = leftSection.
			Foreground(lipgloss.Color("#ff5722")).
			SetString(strings.Join(m.errors, "--"))
	} else if len(m.loadings) > 0 {
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
		leftSection = leftSection.SetString(strings.Join(loadingText, ""))
	} else if m.topicsPage > 0 {
		leftSection = leftSection.SetString(fmt.Sprintf("第%d页", m.topicsPage))
	} else {
		helpKey := consts.AppKeyMap.HelpPage.Help()
		leftSection = leftSection.SetString(fmt.Sprintf("%s %s", helpKey.Key, helpKey.Desc))
	}

	padding := 1
	footer := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftSection.Render(),
		lipgloss.PlaceHorizontal(
			m.screen.Width-lipgloss.Width(leftSection.String())-2*padding,
			lipgloss.Right,
			rightSection.Render(),
		),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Width(m.screen.Width).
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
