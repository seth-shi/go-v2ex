package response

import (
	"fmt"
)

type V2PageResponse struct {
	TotalCount int `json:"total"`
	TotalPages int `json:"pages"`
	CurrPage   int `json:"-"`
}

func (p *V2PageResponse) ToString(def string) string {

	if p.TotalCount <= 0 {
		return def
	}

	return fmt.Sprintf(
		"%d/%d • %d条",
		p.CurrPage,
		p.TotalPages,
		p.TotalCount,
	)
}
