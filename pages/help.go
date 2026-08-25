package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

type helpPage struct{}

type shortcut struct {
	key  string
	desc string
}

func newHelpPage() helpPage { return helpPage{} }

func (m helpPage) Init() tea.Cmd { return nil }

func (m helpPage) OnEntering() (tea.Model, tea.Cmd) {
	g.Session.HideFooter.Store(true)
	return m, nil
}

func (m helpPage) OnLeaving() (tea.Model, tea.Cmd) {
	g.Session.HideFooter.Store(false)
	return m, nil
}

func (m helpPage) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m helpPage) View() string {
	w, h := g.Window.GetSize()
	panelWidth := min(max(w-6, 34), 92)

	navigation := []shortcut{
		{"W/K/↑", "向上选择或滚动"},
		{"S/J/↓", "向下选择或滚动"},
		{"A/H/←", "上一页"},
		{"D/L/→", "下一页"},
		{"Tab", "切换节点"},
		{"Enter", "打开主题 / 加载评论"},
	}
	actions := []shortcut{
		{"R", "解码图片和 Base64"},
		{"F1", "在浏览器中打开"},
		{"U", "检查并安装更新"},
		{"=", "切换底栏显示模式"},
	}
	global := []shortcut{
		{"?", "打开 / 关闭帮助"},
		{"， / `", "打开设置 / 返回"},
		{"Space", "老板键"},
		{"Q", "返回上一页"},
		{"Esc", "退出程序"},
	}

	header := styles.Title.Render("快捷键") + "\n" + styles.Meta.Render("按键随当前页面生效")
	var groups string
	if panelWidth < 60 {
		compactItems := []shortcut{
			{"W/K/↑", "向上选择或滚动"},
			{"S/J/↓", "向下选择或滚动"},
			{"A/H/←", "上一页"},
			{"D/L/→", "下一页"},
			{"Tab", "切换节点"},
			{"Enter", "打开 / 加载"},
			{"R", "解码内容"},
			{"F1", "浏览器打开"},
			{"U", "检查更新"},
			{"=", "切换底栏"},
			{"Space", "老板键"},
			{"?", "帮助"},
			{"， / `", "设置 / 返回"},
			{"Q", "返回"},
			{"Esc", "退出"},
		}
		groups = helpRows(compactItems, panelWidth-6)
	} else {
		gap := 2
		columnWidth := (panelWidth - gap*2 - 6) / 3
		groups = lipgloss.JoinHorizontal(lipgloss.Top,
			helpSection("浏览", navigation, columnWidth),
			strings.Repeat(" ", gap),
			helpSection("操作", actions, columnWidth),
			strings.Repeat(" ", gap),
			helpSection("全局", global, columnWidth),
		)
	}

	note := styles.Meta.Render(ansi.Hardwrap("按 ? 或 Q 返回 · 请求超时时请检查终端代理", panelWidth-6, true))
	content := lipgloss.JoinVertical(lipgloss.Left, header, "", groups, "", note)
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(styles.Theme.Subtle).
		Padding(0, 2).Width(max(panelWidth-6, 1)).
		Render(content)
	return lipgloss.Place(max(w, 1), max(h, 1), lipgloss.Center, lipgloss.Center, panel)
}

func helpSection(title string, items []shortcut, width int) string {
	return styles.Title.Render(title) + "\n\n" + helpRows(items, width)
}

func helpRows(items []shortcut, width int) string {
	var rows strings.Builder
	for i, item := range items {
		keyWidth := min(max(width/3, 6), 8)
		key := lipgloss.NewStyle().Foreground(styles.Theme.Accent).Background(styles.Theme.AccentDim).Bold(true).Align(lipgloss.Center).Width(keyWidth).Render(item.key)
		desc := ansi.TruncateWc(item.desc, max(width-keyWidth-2, 4), "…")
		rows.WriteString(key + " " + styles.Meta.Render(desc))
		if i < len(items)-1 {
			rows.WriteString("\n")
		}
	}
	return rows.String()
}
