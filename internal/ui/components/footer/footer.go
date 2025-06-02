package footer

import (
	"fmt"
	"log"
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
	// 只在 update view 读写, 无需上锁
	// 会自动删除
	loadings map[int]string
	errors   []string
	tips     []string
	// 固定文案, 不会修改 (例如用来显示页码)
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
	case messages.ShowTipsRequest:
		m.leftText = msgType.Text
		return m, nil
	// 消息处理
	case messages.ShowAutoClearTipsRequest:
		log.Println("add clear tips")
		return m, m.addAutoClearTips(msgType.Text)
	case messages.ShiftAutoClearTipsRequest:
		// 删除第一个元素
		m.tips = lo.Slice(m.tips, 1, len(m.tips))
		return m, nil
	case error:
		return m, m.addError(msgType)
	case messages.ShiftErrorRequest:
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

	if len(m.errors) > 0 || len(m.loadings) > 0 || len(m.tips) > 0 || m.leftText != "" {

		if m.leftText != "" {
			leftSection = append(leftSection, m.leftText)
		}

		leftSection = append(leftSection, lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff5722")).
			Render(strings.Join(m.errors, " / ")))

		leftSection = append(leftSection, lipgloss.NewStyle().
			Render(strings.Join(m.tips, " / ")))

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

func (m *Model) addAutoClearTips(text string) tea.Cmd {

	m.tips = append(m.tips, text)
	// 3s 后删除一个
	return tea.Tick(time.Second*3, func(time.Time) tea.Msg {
		return messages.ShiftAutoClearTipsRequest{}
	})
}

func (m *Model) addError(err error) tea.Cmd {

	if err == nil {
		return nil
	}

	m.errors = append(m.errors, err.Error())
	// 3s 后删除一个
	return tea.Tick(time.Second*3, func(time.Time) tea.Msg {
		return messages.ShiftErrorRequest{}
	})
}
