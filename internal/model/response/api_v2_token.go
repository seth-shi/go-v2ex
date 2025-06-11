package response

type V2Token struct {
	Result *v2TokenResult `json:"result"`
}

type v2TokenResult struct {
	Token       string `json:"token"`
	Scope       string `json:"scope"`
	Expiration  int64  `json:"expiration"`
	GoodForDays int    `json:"good_for_days"`
	TotalUsed   int    `json:"total_used"`
	LastUsed    int    `json:"last_used"`
	Created     int64  `json:"created"`
}
