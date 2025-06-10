package response

import (
	"fmt"

	"github.com/seth-shi/go-v2ex/internal/consts"
)

type Page struct {
	// 总记录数
	TotalCount int `json:"total"`
	// 总页数
	TotalPages int `json:"pages"`
	// 当前页
	CurrPage int `json:"currPage"`
}

func (p *Page) ToString() string {
	return fmt.Sprintf(
		"╭─ %d/%d • %d条",
		p.CurrPage,
		p.TotalPages,
		p.TotalCount,
	)
}

func (p *Page) ResetPerPageTo10() *Page {

	if p.TotalCount <= 0 {
		p.TotalPages = 0
		return p
	}

	p.TotalPages = (p.TotalCount + consts.PerPage - 1) / consts.PerPage
	return p
}
