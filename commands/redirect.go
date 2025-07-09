package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevm/bubbleo/navstack"
)

func Redirect(item navstack.NavigationItem) tea.Cmd {
	return navstack.PushNavigationCmd(item)
}

func RedirectPop() tea.Cmd {
	return navstack.PopNavigationCmd()
}
