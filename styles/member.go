package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	MemberPro = lipgloss.NewStyle().
			Background(lipgloss.Color("#323A45")).
			Foreground(lipgloss.Color("#F5F5F6")).
			PaddingLeft(1).
			PaddingRight(1).
			Render("PRO")
	MemberMe = lipgloss.NewStyle().
			Background(lipgloss.Color("#FFF7ED")).
			Foreground(lipgloss.Color("#FB923C")).
			PaddingLeft(1).
			PaddingRight(1).
			Render("YOU")

	MemberOp = lipgloss.NewStyle().
			Background(lipgloss.Color("#ECFDF5")).
			Foreground(lipgloss.Color("#18BC86")).
			PaddingLeft(1).
			PaddingRight(1).
			Render("OP")
)
