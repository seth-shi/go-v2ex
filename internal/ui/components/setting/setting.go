package setting

import (
	"fmt"
	"strings"

	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/ui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#31bdec"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#c2c2c2"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Render("[ 保存 ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("保存"))

	homeFocusedButton = focusedStyle.Render("[ 回到首页 ]")
	homeButton        = fmt.Sprintf("[ %s ]", blurredStyle.Render("回到首页"))

	tipStyle = lipgloss.NewStyle().
			Padding(1, 1, 0, 1)

	formsCount = 4
)

type Model struct {
	focusIndex int
	inputs     []textinput.Model
}

func New() Model {
	m := Model{
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
			t.Prompt = "认证令牌:"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Prompt = "列表节点:"
		}

		m.inputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) RefreshConfig() {
	// 当前不在 body 页, 无法通过消息更新
	if len(m.inputs) > 0 {
		m.inputs[0].SetValue(config.G.Token)
	}

	if len(m.inputs) > 1 {
		m.inputs[1].SetValue(config.G.Nodes)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msgType := msg.(type) {
	case tea.KeyMsg:
		switch msgType.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msgType.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" {

				if m.focusIndex == len(m.inputs) {
					return m, m.saveSettings()
				}

				if m.focusIndex == formsCount-1 {
					return m, messages.Post(
						messages.RedirectTopicsPage{
							Page: 1,
						},
					)
				}
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > formsCount {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = formsCount - 1
			}

			// 更新表单的值
			cmds := make([]tea.Cmd, formsCount)
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

func (m Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) saveSettings() tea.Cmd {
	if len(m.inputs) > 0 {
		config.G.Token = strings.TrimSpace(m.inputs[0].Value())
	}

	if len(m.inputs) > 1 {
		config.G.Nodes = strings.TrimSpace(m.inputs[1].Value())
	}

	api.V2ex.RefreshConfig()
	return config.SaveToFile("配置保存成功")
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(styles.Err.PaddingLeft(1).Render("tab 切换表单, 回车确认(如有请求超时, 请设置 clash 全局代理, 或者复制代理环境变量到终端执行)"))
	b.WriteString("\n")
	text := fmt.Sprintf("配置文件路径: %s", config.SavePath())
	b.WriteString(styles.Bold.PaddingLeft(1).Render(text))

	if len(m.inputs) > 0 {
		text := fmt.Sprintf(
			"\n%s\n%s",
			"点此创建秘钥: https://www.v2ex.com/settings/tokens",
			m.inputs[0].View(),
		)
		b.WriteString(tipStyle.Render(text))
	}

	if len(m.inputs) > 1 {
		text := fmt.Sprintf(
			"\n%s\n%s",
			"所有分类此处查看: https://v2ex.com/planes (多个分类使用英文逗号隔开, URL 上的 https://v2ex.com/go/{name})",
			m.inputs[1].View(),
		)
		b.WriteString(tipStyle.Render(text))
	}

	btn1 := &blurredButton
	// 最后一个 input
	if m.focusIndex == len(m.inputs) {
		btn1 = &focusedButton
	}

	btn2 := &homeButton
	// 最后一个 input
	if m.focusIndex == formsCount-1 {
		btn2 = &homeFocusedButton
	}

	b.WriteString(tipStyle.Render(fmt.Sprintf("\n%s    %s\n", *btn1, *btn2)))
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(config.Screen.Width - 2).
		Height(config.Screen.Height - 4).
		Padding(1).
		Render(b.String())
}
