package config

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mcuadros/go-defaults"
)

var (
	Screen = newScreen()
)

type screen struct {
	Height  int
	Width   int
	Padding int `default:"1"`
}

func newScreen() screen {
	var s screen
	defaults.SetDefaults(&s)
	return s
}

func (s *screen) Sync(msg tea.WindowSizeMsg) {
	s.Height = msg.Height
	s.Width = msg.Width
}

func (s *screen) GetContentWidth() int {
	// left + right padding
	return s.Width - 2
}
