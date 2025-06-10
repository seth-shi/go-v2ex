package footer

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/ui/styles"
)

var (
	rightText = fmt.Sprintf("%s@%s Powered by seth-shi", consts.AppName, consts.AppVersion)
)

type Model struct {
	// 只在 update view 读写, 无需上锁, 会自动删除
	loadings map[int]string
	errors   []string
	tips     []string
	// 固定文案, 不会修改 (例如用来显示页码)
	leftText string
	helpText string
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
	case messages.ShowAlertRequest:
		m.leftText = msgType.Text
		m.helpText = msgType.Help
		return m, nil
	// 消息处理
	case messages.ShowToastRequest:
		m.tips = append(m.tips, msgType.Text)
		return m, tea.Tick(
			time.Second*3, func(time.Time) tea.Msg {
				return messages.ShiftToastRequest{}
			},
		)
	case error:
		if msgType == nil {
			return m, nil
		}
		m.errors = append(m.errors, msgType.Error())
		return m, tea.Tick(
			time.Second*3, func(time.Time) tea.Msg {
				return messages.ShiftErrorRequest{}
			},
		)
	// 有定时器触发这个删除
	case messages.ShiftToastRequest:
		m.tips = lo.Slice(m.tips, 1, len(m.tips))
		return m, nil
	// 有定时器触发这个删除
	case messages.ShiftErrorRequest:
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
		showFooter  = config.G.ShowFooter()
		showHelp    = config.G.ShowHelp()
		showLimit   = config.G.ShowLimit()
		screenWidth = config.Screen.Width
		padding     = config.Screen.Padding
		footer      strings.Builder
	)

	// 外部传入的 text 优先级最高显示
	// 然后显示错误消息(n 秒后自动删除)
	// 然后显示提示消息(n 秒后自动删除)
	// 显示 loading 消息(需要调用方手动删除)
	if m.leftText != "" && showFooter {
		footer.WriteString(styles.Hint.Render(m.leftText))
	}
	// 错误 && 加载 && 提示不管用户显示不显示页脚都不影响
	if len(m.errors) > 0 {
		footer.WriteString(" ")
		footer.WriteString(styles.Err.Render(strings.Join(m.errors, " ")))
	}
	if len(m.tips) > 0 {
		footer.WriteString(" ")
		footer.WriteString(styles.Hint.Render(strings.Join(m.tips, " ")))
	}

	if len(m.loadings) > 0 {
		footer.WriteString(" ")
		footer.WriteString(m.getLoadingText())
	}

	if m.helpText != "" && showHelp {
		footer.WriteString(" ")
		footer.WriteString(styles.Hint.Render(m.helpText))
	}

	if showFooter {
		currentContent := footer.String()
		paddingLeft := screenWidth - lipgloss.Width(currentContent) - 2*padding
		footer.Reset()
		footer.WriteString(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				currentContent,
				lipgloss.PlaceHorizontal(
					paddingLeft,
					lipgloss.Right,
					styles.Hint.Render(rightText),
				),
			),
		)

		if showLimit {
			rate := api.V2ex.GetLimitRate()
			borderWidth := int(math.Round(float64(screenWidth) * rate))
			footer.WriteString("\n")
			footer.WriteString(strings.Repeat("♡", max(0, borderWidth)))
			footer.WriteString(strings.Repeat("_", max(0, screenWidth-borderWidth)))
		}
	}

	return styles.
		Hint.
		Width(screenWidth).
		Render(footer.String())
}

func (m Model) getLoadingText() string {
	// loadings 是一个 map
	var (
		loadingKeys = lo.Keys(m.loadings)
		loadingIcon = m.spinner.View()
		loadingText strings.Builder
	)
	slices.Sort(loadingKeys)

	lo.ForEach(
		loadingKeys, func(key int, index int) {
			loadingText.WriteString(" ")
			loadingText.WriteString(loadingIcon)
			loadingText.WriteString(" ")
			loadingText.WriteString(m.loadings[key])
		},
	)
	return styles.Hint.Render(loadingText.String())
}
