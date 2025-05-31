package types

import tea "github.com/charmbracelet/bubbletea"

type ScreenSize struct {
	Height int
	Width  int
}

func (s *ScreenSize) Sync(msg tea.WindowSizeMsg) {
	s.Height = msg.Height
	s.Width = msg.Width
}
