package response

import (
	"fmt"
)

type V2PageResponse struct {
	TotalCount int `json:"total"`
	TotalPages int `json:"pages"`
}

func (p *V2PageResponse) ToString(currPage int) string {
	return fmt.Sprintf(
		"%d/%d • %d条",
		currPage,
		p.TotalPages,
		p.TotalCount,
	)
}
