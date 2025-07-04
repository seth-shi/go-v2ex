package components

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/commands"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/ui/styles"
)

type FooterComponents struct {
	// 只在 update view 读写, 无需上锁, 会自动删除
	loadings map[int]string
	// 固定文案, 不会修改 (例如用来显示页码)
	leftText   string
	helpText   string
	spinner    spinner.Model
	appVersion string

	statusBar    statusbar.Model
	hiddenFooter bool
}

func NewFooter(appVersion string) FooterComponents {

	sb := statusbar.New(
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F25D94"},
		},
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#3c3836"},
		},
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#A550DF", Dark: "#A550DF"},
		},
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#6124DF", Dark: "#6124DF"},
		},
	)
	return FooterComponents{
		// 最大加载数限定
		loadings:   make(map[int]string, 10),
		spinner:    spinner.New(spinner.WithSpinner(spinner.Points)),
		statusBar:  sb,
		appVersion: appVersion,
	}
}

func (m FooterComponents) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
	)
}

func (m FooterComponents) Update(msg tea.Msg) (FooterComponents, tea.Cmd) {

	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	m.statusBar, cmd = m.statusBar.Update(msg)
	cmds = append(cmds)

	switch msg := msg.(type) {
	case messages.FooterStatusMessage:
		m.hiddenFooter = msg.HiddenFooter
	// 把错误转到到另一个消息里
	case error:
		cmds = append(cmds, commands.AlertError(msg))
	case messages.StartLoading:
		m.loadings[msg.ID] = msg.Text
	case messages.EndLoading:
		delete(m.loadings, msg.ID)
	case messages.ShowAlertRequest:
		m.leftText = msg.Text
		m.helpText = msg.Help
	// 消息处理
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.UpgradeApp):
			cmds = append(cmds, commands.UpgradeApp(m.appVersion))
		case key.Matches(msg, consts.AppKeyMap.SwitchShowMode):
			config.G.SwitchShowMode()
			// 保存配置
			return m, tea.Batch(
				// messages.ErrorOrToast(commands.SaveToFile(config.G), ""),
				messages.Post(messages.ShowToastRequest{Text: config.G.GetShowModeText()}),
			)
		}
	case messages.UpgradeStateMessage:
		m.leftText = msg.State.Text()
		cmds = append(cmds, tea.Tick(time.Second, commands.CheckDownloadProcessMessages(msg.State)))
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	}

	return m, tea.Batch(cmds...)
}

func (m FooterComponents) View() string {

	if m.hiddenFooter {
		return ""
	}

	// 1. 显示分页信息
	// 2. 显示加载动画, help 信息
	// 3. 版本号
	// 4 name
	var (
		first  = m.GetFirstColumnContent()
		second = m.GetSecondColumnContent()
		third  = m.GetThirdColumnContent()
		fourth = m.GetFourthColumContent()
	)
	m.statusBar.SetContent(first, second, third, fourth)
	return m.statusBar.View()
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
					styles.Hint.Render(fmt.Sprintf("%s@%s Powered by seth-shi", consts.AppName, m.appVersion)),
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

func (m FooterComponents) GetFirstColumnContent() string {
	return fmt.Sprintf("%s@%s", consts.AppName, m.appVersion)
}

func (m FooterComponents) GetSecondColumnContent() string {
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

func (m FooterComponents) GetThirdColumnContent() string {
	return fmt.Sprintf("%s@%s", consts.AppName, m.appVersion)
}

func (m FooterComponents) GetFourthColumContent() string {
	return fmt.Sprintf("Powered by %s", consts.AppOwner)
}
