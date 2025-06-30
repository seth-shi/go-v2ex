package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	BoldBorder = lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder())
	Border     = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
)
