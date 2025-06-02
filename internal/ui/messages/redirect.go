package messages

import tea "github.com/charmbracelet/bubbletea"

type RedirectPageRequest struct {
	Page tea.Model
}

type RedirectDetailRequest struct {
	Id int64
}
