package pages

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mistakenelf/teacup/statusbar"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/api"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/nav"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

type FooterComponents struct {
	// 只在 update view 读写, 无需上锁, 会自动删除
	loadings map[int]string
	// 固定文案, 不会修改 (例如用来显示页码)
	secondText string
	firstText  string
	spinner    spinner.Model
	appVersion string

	statusBar statusbar.Model
}

func NewFooter(appVersion string) FooterComponents {

	sb := statusbar.New(
		statusbar.ColorConfig{
			// 主区域
			Foreground: lipgloss.AdaptiveColor{Dark: "#FFFFFF", Light: "#FFFFFF"},
			Background: lipgloss.AdaptiveColor{Light: "#005FB8", Dark: "#005FB8"},
		},
		statusbar.ColorConfig{
			// 辅助区1
			Foreground: lipgloss.AdaptiveColor{Dark: "#999999", Light: "#999999"},
			Background: lipgloss.AdaptiveColor{Light: "#F8F8F8", Dark: "#F8F8F8"},
		},
		statusbar.ColorConfig{
			// 辅助区2
			Foreground: lipgloss.AdaptiveColor{Dark: "#EEEEEE", Light: "#EEEEEE"},
			Background: lipgloss.AdaptiveColor{Light: "#636e72", Dark: "#636e72"},
		},
		statusbar.ColorConfig{
			// 强调区
			Foreground: lipgloss.AdaptiveColor{Dark: "#FFFFFF", Light: "#FFFFFF"},
			Background: lipgloss.AdaptiveColor{Light: "#005FB8", Dark: "#005FB8"},
		},
	)

	sb.SetContent(
		"",
		"",
		"?查看帮助",
		fmt.Sprintf("%s[%s]@%s", consts.AppName, appVersion, consts.AppOwner),
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
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case messages.CheckUpgradeAppRequest:
		cmds = append(cmds, commands.CheckAppHasNewVersion(m.appVersion))
	// 把错误转到到另一个消息里
	case error:
		cmds = append(cmds, commands.AlertError(msg))
	case messages.StartLoading:
		m.loadings[msg.ID] = msg.Text
	case messages.EndLoading:
		delete(m.loadings, msg.ID)
	case messages.ShowStatusBarTextRequest:
		m.secondText = msg.HelpText
		m.firstText = msg.FirstText
		m.statusBar.FirstColumn = msg.FirstText
	// 不直接发消息, 因为 msg需要一个延迟, 代理转发
	case messages.ProxyShowToastRequest:
		cmds = append(cmds, commands.AlertInfo(msg.Text))
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, consts.AppKeyMap.UpgradeApp):
			cmds = append(cmds, commands.UpgradeApp(m.appVersion))
		case key.Matches(msg, consts.AppKeyMap.SwitchShowMode):
			cmds = append(cmds, m.onSwitchShowMode())
		}
	case messages.UpgradeStateMessage:
		m.statusBar.FirstColumn = msg.State.Text()
		cmds = append(cmds, tea.Tick(time.Second, commands.CheckDownloadProcessMessages(msg.State)))
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
func (m FooterComponents) onSwitchShowMode() tea.Cmd {

	return func() tea.Msg {

		err := g.Config.Save(
			func(conf *model.FileConfig) {
				conf.SwitchShowMode()
			},
		)

		return messages.ErrorOrToast(err, g.Config.Get().GetShowModeText())
	}
}
func (m FooterComponents) View() string {

	// 全局改写
	if g.Session.HideFooter.Load() {
		return ""
	}

	var (
		conf = g.Config.Get()
		w, _ = g.Window.GetSize()
	)

	if !conf.ShowFooter() {
		return ""
	}

	var name string
	ref := reflect.TypeOf(nav.CurrentPage())
	if ref == nil {
		name = "nil"
	} else {
		name = ref.Name()
	}
	left := m.firstText
	if loading := m.GetSecondColumnContent(); lipgloss.Width(loading) > 0 {
		left = loading
	}
	if left == "" {
		left = pageLabel(name)
	}

	if conf.ShowLimit() && w >= 90 {
		left += fmt.Sprintf("  ·  API %d%%", int(math.Round(api.V2ex.GetLimitRate()*100)))
	}
	left = ansi.TruncateWc(left, max(w/3, 8), "…")
	left = lipgloss.NewStyle().Foreground(styles.Theme.Text).Bold(true).Render(left)
	hints := footerHints(name, w)
	gap := max(w-lipgloss.Width(left)-lipgloss.Width(hints)-2, 1)
	line := " " + left + strings.Repeat(" ", gap) + hints
	return lipgloss.NewStyle().
		Background(styles.Theme.Surface).
		Width(max(w, 1)).
		Render(ansi.TruncateWc(line, max(w, 1), "…"))

}

func footerHints(page string, width int) string {
	keyText := func(v string) string { return styles.Key.Render(v) }
	switch page {
	case "topicPage":
		if width < 54 {
			return fmt.Sprintf("%s  %s 打开  %s 设置", keyText("↑↓"), keyText("Enter"), keyText(","))
		}
		if width < 76 {
			return fmt.Sprintf("%s 选择  %s 打开  %s 节点  %s 设置", keyText("↑↓"), keyText("Enter"), keyText("Tab"), keyText(","))
		}
		return fmt.Sprintf("%s 选择  %s 打开  %s 翻页  %s 节点  %s 设置  %s 帮助", keyText("↑↓"), keyText("Enter"), keyText("←→"), keyText("Tab"), keyText(","), keyText("?"))
	case "detailPage":
		if width < 64 {
			return fmt.Sprintf("%s 滚动  %s 更多  %s 返回", keyText("↑↓"), keyText("Enter"), keyText("Q"))
		}
		return fmt.Sprintf("%s 滚动   %s 加载评论   %s 浏览器打开   %s 返回", keyText("↑↓"), keyText("Enter"), keyText("F1"), keyText("Q"))
	default:
		if width < 54 {
			return fmt.Sprintf("%s 返回  %s 设置  %s 退出", keyText("Q"), keyText("`"), keyText("Esc"))
		}
		return fmt.Sprintf("%s 返回   %s 设置   %s 帮助   %s 退出", keyText("Q"), keyText("`"), keyText("?"), keyText("Esc"))
	}
}

func pageLabel(page string) string {
	switch page {
	case "topicPage":
		return "主题"
	case "detailPage":
		return "详情"
	case "splashPage":
		return "正在启动"
	default:
		return "Go V2EX"
	}
}

func (m FooterComponents) GetSecondColumnContent() string {
	// loadings 是一个 map
	var (
		loadingKeys = lo.Keys(m.loadings)
		loadingIcon = m.spinner.View()
		loadingText strings.Builder
	)
	slices.Sort(loadingKeys)
	if len(loadingKeys) == 0 {
		return styles.Hint.Render(m.secondText)
	}
	loadingText.WriteString(loadingIcon)
	loadingText.WriteString(" ")
	loadingText.WriteString(m.loadings[loadingKeys[0]])
	if len(loadingKeys) > 1 {
		loadingText.WriteString(fmt.Sprintf("  · 另有 %d 项", len(loadingKeys)-1))
	}
	return styles.Hint.Render(loadingText.String())
}
