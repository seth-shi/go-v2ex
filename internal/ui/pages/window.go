package pages

import (
	tea "github.com/charmbracelet/bubbletea"
)

type windowPage struct {
	w int
	h int
}

func (m windowPage) Update(msg tea.Msg) windowPage {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w = msg.Width
		m.h = msg.Height
	}
	return m
}
