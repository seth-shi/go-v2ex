package response

type MemberResult struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}

type NodeInfoResult struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
}
