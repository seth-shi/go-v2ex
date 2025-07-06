package pages

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/ui/styles"
)

func loadingView(w, h int, title string) string {
	return styles.Err.
		Align(lipgloss.Center).
		PaddingTop(max(h/4, 2)).
		Bold(true).
		Width(w).
		Render(title)
}
