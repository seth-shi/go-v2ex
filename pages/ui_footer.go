package pages

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
	"github.com/seth-shi/go-v2ex/v2/api"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

type FooterComponents struct {
	// 只在 update view 读写, 无需上锁, 会自动删除
	loadings map[int]string
	// 固定文案, 不会修改 (例如用来显示页码)
	secondText string
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
	cmds = append(cmds)

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
		conf       = g.Config.Get()
		content    strings.Builder
		w, _       = g.Window.GetSize()
		secondText = m.GetSecondColumnContent()
	)

	if !conf.ShowFooter() {
		return ""
	}

	if conf.ShowLimit() {

		rate := api.V2ex.GetLimitRate()
		borderWidth := int(math.Round(float64(w) * rate))
		content.WriteString(strings.Repeat("♡", max(0, borderWidth)))
		content.WriteString(strings.Repeat("_", max(0, w-borderWidth)))
		content.WriteString("\n")
	}

	// 这一列有 loading 动画, 需要实时计算
	m.statusBar.SecondColumn = secondText
	content.WriteString(m.statusBar.View())

	return content.String()

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

	return styles.Hint.Render(loadingText.String(), m.secondText)
}
