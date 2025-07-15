package response

import (
	"fmt"

	"github.com/seth-shi/go-v2ex/consts"
	"github.com/seth-shi/go-v2ex/pkg"
)

type PerTenPageInfo struct {
	// 总记录数
	TotalCount int `json:"total"`
	// 当前页
	CurrPage int `json:"currPage"`
}

func (p *PerTenPageInfo) ToString() string {
	return fmt.Sprintf(
		"%d/%d • %d条",
		p.CurrPage,
		p.TotalPage(),
		p.TotalCount,
	)
}

func (p *PerTenPageInfo) TotalPage() int {
	return pkg.TotalPages(p.TotalCount, consts.PerPage)
}
