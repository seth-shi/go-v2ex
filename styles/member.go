package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	MemberPro = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#323A45", Dark: "#34404C"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#F5F5F6", Dark: "#E6EDF3"}).
			PaddingLeft(1).
			PaddingRight(1).
			Render("PRO")
	MemberMe = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#FFF0DC", Dark: "#4B2F16"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FDBA74"}).
			PaddingLeft(1).
			PaddingRight(1).
			Render("YOU")

	MemberOp = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#E6F8F0", Dark: "#143C30"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#13795B", Dark: "#5AD7A0"}).
			PaddingLeft(1).
			PaddingRight(1).
			Render("OP")
)
