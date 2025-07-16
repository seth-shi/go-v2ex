package pages

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

func loadingView(title string) string {

	var (
		width, height = g.Window.GetSize()
	)

	return styles.Bold.
		Align(lipgloss.Center).
		PaddingTop(max(height/4, 2)).
		Bold(true).
		Width(width).
		Render(title)
}
