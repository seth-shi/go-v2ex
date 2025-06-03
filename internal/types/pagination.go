package types

type Pagination struct {
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
	Pages   int `json:"pages"`
}
