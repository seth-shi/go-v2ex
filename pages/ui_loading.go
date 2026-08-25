package pages

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

func loadingView(title string) string {
	width, height := g.Window.GetSize()
	if !g.Session.HideFooter.Load() && g.Config.Get().ShowFooter() {
		height--
	}
	return loadingViewWithin(title, width, height)
}

func loadingViewWithin(title string, width, height int) string {
	brand := styles.Active.Bold(true).Render("◆  GO V2EX")
	status := styles.Title.Render(title)
	content := lipgloss.JoinVertical(lipgloss.Center, brand, "", status, styles.Meta.Render("正在加载，请稍候…"))
	return lipgloss.Place(max(width, 1), max(height, 1), lipgloss.Center, lipgloss.Center, content)
}
