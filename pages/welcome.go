package pages

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

func welcomeView() string {
	w, h := g.Window.GetSize()
	compact := w < 50 || h < 27
	cardWidth := min(max(w-8, 30), 62)
	rows := []string{
		welcomeAction("Enter", "直接浏览，无需令牌", cardWidth),
		welcomeAction(",", "配置令牌，启用完整 API", cardWidth),
		welcomeAction("↑↓", "浏览和选择主题", cardWidth),
		welcomeAction("Tab", "切换主题节点", cardWidth),
		welcomeAction("?", "查看全部快捷键", cardWidth),
	}
	brand := styles.Active.Bold(true).Render("◆  GO V2EX")
	title := styles.Title.Render("欢迎使用终端版 V2EX")
	subtitle := styles.Meta.Render("令牌是可选项，不配置也能浏览主题和评论")
	button := lipgloss.NewStyle().Foreground(styles.Theme.OnAccent).Background(styles.Theme.Accent).Bold(true).Padding(0, 3).Render("Enter  直接浏览")
	if compact {
		rows = []string{
			welcomeAction("Enter", "直接浏览", cardWidth),
			welcomeAction(",", "配置令牌", cardWidth),
			welcomeAction("?", "查看帮助", cardWidth),
		}
	}
	body := lipgloss.JoinVertical(lipgloss.Center, brand, "", title, subtitle, "", strings.Join(rows, "\n"), "", button)
	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.Theme.Subtle).Align(lipgloss.Center).Padding(1, 2).Width(max(cardWidth-6, 1)).Render(body)
	return lipgloss.Place(max(w, 1), max(h, 1), lipgloss.Center, lipgloss.Center, card)
}

func welcomeAction(keyText, description string, width int) string {
	keyStyle := lipgloss.NewStyle().Foreground(styles.Theme.Accent).Background(styles.Theme.AccentDim).Bold(true).Align(lipgloss.Center).Width(9)
	line := keyStyle.Render(keyText) + "  " + styles.Meta.Render(description)
	return lipgloss.NewStyle().Width(max(width-8, 1)).Align(lipgloss.Left).Render(line)
}
