package setting

import (
	"fmt"
	"log"
	"strings"

	"github.com/seth-shi/go-v2ex/internal/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#31bdec"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#c2c2c2"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Render("[ 保存 ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("保存"))

	tipStyle = lipgloss.NewStyle().
			Padding(1, 1, 0, 1)
)

type Model struct {
	focusIndex int
	config     types.FileConfig
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

func (m Model) SetConfig(cfg types.FileConfig) {
	m.config = cfg
	if len(m.inputs) > 0 {
		m.inputs[0].SetValue(m.config.Token)
	}

	if len(m.inputs) > 1 {
		m.inputs[1].SetValue(m.config.Nodes)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, m.saveSettings()
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

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

			// TODO 清楚所有错误
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
		m.config.Token = strings.TrimSpace(m.inputs[0].Value())
	}

	if len(m.inputs) > 1 {
		m.config.Nodes = strings.TrimSpace(m.inputs[1].Value())
	}

	// 保存数据
	if err := m.config.SaveToFile(); err != nil {
		// 停留在此页面
		return messages.Post(err)
	}

	return messages.Post(messages.SettingSaveResult{Config: m.config})
}

func (m Model) View() string {
	var b strings.Builder

	log.Println("setting")
	log.Println(m.config)

	text := fmt.Sprintf(
		"配置文件路径:%s",
		m.config.ConfigPath(),
	)
	b.WriteString(tipStyle.Render(text))

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
			"所有分类此处查看: https://www.v2ex.com/api/nodes/all.json (多个分类使用英文逗号隔开)",
			m.inputs[1].View(),
		)
		b.WriteString(tipStyle.Render(text))
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}

	b.WriteString(tipStyle.Render(fmt.Sprintf("\n%s\n", *button)))
	return b.String()
}
