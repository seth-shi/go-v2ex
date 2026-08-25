package styles

import "github.com/charmbracelet/lipgloss"

// Theme keeps the application palette in one place. Adaptive colors make the
// UI readable in both dark and light terminals without every page knowing
// which palette is active.
var Theme = struct {
	Accent    lipgloss.AdaptiveColor
	OnAccent  lipgloss.AdaptiveColor
	AccentDim lipgloss.AdaptiveColor
	Text      lipgloss.AdaptiveColor
	Muted     lipgloss.AdaptiveColor
	Subtle    lipgloss.AdaptiveColor
	Surface   lipgloss.AdaptiveColor
	Selected  lipgloss.AdaptiveColor
	Success   lipgloss.AdaptiveColor
	Warning   lipgloss.AdaptiveColor
	Danger    lipgloss.AdaptiveColor
}{
	Accent:    lipgloss.AdaptiveColor{Light: "#0067C0", Dark: "#64B5F6"},
	OnAccent:  lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#0B1722"},
	AccentDim: lipgloss.AdaptiveColor{Light: "#DCEEFF", Dark: "#17324A"},
	Text:      lipgloss.AdaptiveColor{Light: "#18212B", Dark: "#E6EDF3"},
	Muted:     lipgloss.AdaptiveColor{Light: "#66717E", Dark: "#8B98A5"},
	Subtle:    lipgloss.AdaptiveColor{Light: "#D7DEE5", Dark: "#34404C"},
	Surface:   lipgloss.AdaptiveColor{Light: "#F5F8FA", Dark: "#17212B"},
	Selected:  lipgloss.AdaptiveColor{Light: "#E8F3FC", Dark: "#1B3042"},
	Success:   lipgloss.AdaptiveColor{Light: "#13795B", Dark: "#5AD7A0"},
	Warning:   lipgloss.AdaptiveColor{Light: "#A15C00", Dark: "#F5B942"},
	Danger:    lipgloss.AdaptiveColor{Light: "#C4384B", Dark: "#FF7B8B"},
}

var (
	Title       = lipgloss.NewStyle().Foreground(Theme.Text).Bold(true)
	Meta        = lipgloss.NewStyle().Foreground(Theme.Muted)
	Divider     = lipgloss.NewStyle().Foreground(Theme.Subtle)
	SelectedRow = lipgloss.NewStyle().Background(Theme.Selected)
	Key         = lipgloss.NewStyle().Foreground(Theme.Accent).Bold(true)
)
