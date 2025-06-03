package messages

import tea "github.com/charmbracelet/bubbletea"

type RedirectPageRequest struct {
	ContentModel tea.Model
}

type RedirectDetailRequest struct {
	Id int64
}

type RedirectTopicsPage struct {
	Page int
}
