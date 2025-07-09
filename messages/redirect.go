package messages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevm/bubbleo/navstack"
)

func Redirect(item navstack.NavigationItem) tea.Msg {
	return navstack.PushNavigation{Item: item}
}
