package config

import tea "github.com/charmbracelet/bubbletea"

var (
	Screen screen
)

type screen struct {
	Height int
	Width  int
}

func (s *screen) Sync(msg tea.WindowSizeMsg) {
	s.Height = msg.Height
	s.Width = msg.Width
}

func (s *screen) GetContentWidth() int {
	// left + right padding
	return s.Width - 2
}
