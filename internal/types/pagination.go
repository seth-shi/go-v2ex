package types

import (
	"fmt"
)

type Pagination struct {
	PerPage  int `json:"per_page"`
	Total    int `json:"total"`
	Pages    int `json:"pages"`
	CurrPage int `json:"currPage"`
}

func (p *Pagination) ResetPages(perPage, total int) *Pagination {
	p.PerPage = perPage
	p.Total = total
	if p.Total <= 0 {
		p.Pages = 0
		return p
	}

	p.Pages = (p.Total + p.PerPage - 1) / p.PerPage
	return p
}

func (p Pagination) ToString(ext string) string {
	return fmt.Sprintf(
		"%d / %d (%d) %s",
		p.CurrPage,
		p.Pages,
		p.Total,
		ext,
	)
}
