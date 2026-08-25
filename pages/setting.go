package pages

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/pkg/browser"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/reflow/wrap"
)

const (
	tokenSettingsURL = "https://www.v2ex.com/settings/tokens"
	nodeDirectoryURL = "https://www.v2ex.com/planes"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(styles.Theme.Accent)
	blurredStyle = lipgloss.NewStyle().Foreground(styles.Theme.Muted)
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
)

type settingPage struct {
	focusIndex int
	inputs     []textinput.Model
	loaded     bool
}

func newSettingPage() settingPage {
	m := settingPage{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 500

		switch i {
		case 0:
			t.Placeholder = ""
			t.Prompt = ""
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Prompt = ""
			t.Placeholder = "programmer,qna,jobs"
		}

		m.inputs[i] = t
	}

	return m
}

func (m settingPage) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
	)
}

func (m settingPage) OnEntering() (tea.Model, tea.Cmd) {

	// 屏蔽 Q 返回键盘
	consts.AppKeyMap.KeyQ.SetEnabled(false)
	consts.AppKeyMap.UpgradeApp.SetEnabled(false)
	g.Session.HideFooter.Store(true)

	return m, commands.Post(
		messages.LoadConfigResult{
			Result: g.Config.Get(),
		},
	)
}

func (m settingPage) OnLeaving() (tea.Model, tea.Cmd) {
	g.Session.HideFooter.Store(false)
	consts.AppKeyMap.UpgradeApp.SetEnabled(true)
	consts.AppKeyMap.KeyQ.SetEnabled(true)

	return m, nil
}

func (m settingPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for i := range m.inputs {
			m.inputs[i].Width = max(min(msg.Width-10, 72), 16)
		}
	case messages.LoadConfigResult:
		m.loaded = true
		m.inputs[0].SetValue(msg.Result.Token)
		m.inputs[1].SetValue(msg.Result.MyNodes)
	case tea.KeyMsg:

		if key.Matches(msg, consts.AppKeyMap.F1) {
			return m, func() tea.Msg {
				return browser.OpenFile(model.ConfigPath())
			}
		}

		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" {
				if m.focusIndex == len(m.inputs) {
					return m, m.saveSettings()
				}
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.inputs)+1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			// 更新表单的值
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m settingPage) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m settingPage) saveSettings() tea.Cmd {

	var (
		token   = strings.TrimSpace(lo.NthOrEmpty(m.inputs, 0).Value())
		myNodes = strings.TrimSpace(lo.NthOrEmpty(m.inputs, 1).Value())
	)

	return func() tea.Msg {
		err := g.Config.Save(
			func(conf *model.FileConfig) {
				conf.Token = token
				conf.MyNodes = myNodes
			},
		)
		return messages.ErrorOrToast(err, "保存配置成功")
	}
}

func (m settingPage) View() string {
	var (
		w, h = g.Window.GetSize()
	)

	if !m.loaded {
		return loadingView("载入配置中...")
	}

	panelWidth := min(max(w-8, 32), 76)
	inputWidth := max(panelWidth-8, 16)
	compact := h < 30 || w < 50
	for i := range m.inputs {
		m.inputs[i].Width = inputWidth
	}

	header := styles.Title.Render("偏好设置") + "\n" + styles.Meta.Render("连接 V2EX，并选择你想关注的节点")
	tokenHelp := "可选；配置后启用完整 V2 API 和个人信息，不配置也能公开浏览\n" + terminalLink(tokenSettingsURL, "前往 v2ex.com/settings/tokens 创建令牌 ↗")
	nodeHelp := wrap.String("多个节点使用英文逗号分隔，例如 programmer,qna,jobs", inputWidth) + "\n" + terminalLink(nodeDirectoryURL, "前往 v2ex.com/planes 查看全部节点 ↗")
	if compact {
		header = styles.Title.Render("偏好设置")
		tokenHelp = terminalLink(tokenSettingsURL, "创建令牌 ↗")
		nodeLink := terminalLink(nodeDirectoryURL, "查看节点列表 ↗")
		if w >= 54 {
			tokenHelp = terminalLink(tokenSettingsURL, "在 v2ex.com/settings/tokens 创建令牌 ↗")
			nodeLink = terminalLink(nodeDirectoryURL, "前往 v2ex.com/planes 查看全部节点 ↗")
		}
		nodeHelp = wrap.String("英文逗号分隔，如 programmer,qna", inputWidth) + "\n" + nodeLink
	}
	tokenField := settingField("认证令牌", tokenHelp, m.inputs[0].View(), m.focusIndex == 0, panelWidth)
	nodeField := settingField("我的节点", nodeHelp, m.inputs[1].View(), m.focusIndex == 1, panelWidth)

	button := lipgloss.NewStyle().Padding(0, 3).Bold(true).Render("保存设置")
	if m.focusIndex == len(m.inputs) {
		button = lipgloss.NewStyle().Foreground(styles.Theme.OnAccent).Background(styles.Theme.Accent).Padding(0, 3).Bold(true).Render("保存设置")
	} else {
		button = lipgloss.NewStyle().Foreground(styles.Theme.Muted).Background(styles.Theme.Surface).Padding(0, 3).Render("保存设置")
	}
	actionHint := "Tab 切换  ·  Enter 确认  ·  F1 配置文件  ·  ` 返回"
	var actions string
	if compact {
		actionHint = "Tab 切换  ·  ` 返回"
		actions = lipgloss.JoinVertical(lipgloss.Center, button, styles.Meta.Render(actionHint))
	} else {
		actions = lipgloss.JoinHorizontal(lipgloss.Center, button, styles.Meta.MarginLeft(2).Render(actionHint))
	}
	pathText := styles.Meta.Render(ansi.TruncateWc("配置文件  "+model.ConfigPath(), panelWidth, "…"))

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", tokenField, "", nodeField, "", actions)
	if !compact {
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", pathText)
	}
	content = lipgloss.NewStyle().Width(panelWidth).Render(content)
	return lipgloss.Place(max(w, 1), max(h, 1), lipgloss.Center, lipgloss.Center, content)
}

func settingField(title, description, input string, active bool, width int) string {
	borderColor := styles.Theme.Subtle
	marker := styles.Meta.Render("○")
	if active {
		borderColor = styles.Theme.Accent
		marker = styles.Active.Render("●")
	}
	heading := marker + "  " + styles.Title.Render(title)
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).
		Padding(0, 1).
		Width(max(width-4, 1)).
		Render(input)
	return heading + "\n" + styles.Meta.Render(description) + "\n" + inputBox
}

func terminalLink(url, label string) string {
	return ansi.SetHyperlink(url) + styles.Active.Underline(true).Render(label) + ansi.ResetHyperlink()
}
