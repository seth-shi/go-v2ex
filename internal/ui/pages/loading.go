package pages

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/ui/styles"
)

func loadingView(w, h int, title string) string {
	return styles.Err.
		Align(lipgloss.Center).
		PaddingTop(max(h/2, 10)).
		Bold(true).
		Width(max(w, 20)).
		Render(title)
}
