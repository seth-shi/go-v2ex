package response

type NodeInfoResult struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Title  string `json:"title"`
	Topics int    `json:"topics"`
}
