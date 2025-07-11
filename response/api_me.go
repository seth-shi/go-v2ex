package response

type MeResponse struct {
	Success bool     `json:"success"`
	Result  MeResult `json:"result"`
}

type MeResult struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	Url          string `json:"url"`
	Website      string `json:"website"`
	Twitter      string `json:"twitter"`
	Psn          string `json:"psn"`
	Github       string `json:"github"`
	Btc          string `json:"btc"`
	Location     string `json:"location"`
	Tagline      string `json:"tagline"`
	Bio          string `json:"bio"`
	Created      int    `json:"created"`
	LastModified int    `json:"last_modified"`
	Pro          int    `json:"pro"`
}
