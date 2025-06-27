package response

type GroupTopic struct {
	Items      []TopicResult
	Pagination *PerTenPageInfo
}

type TopicResult struct {
	Id          int64  `json:"id"`
	Node        string `json:"node"`
	Title       string `json:"title"`
	Member      string `json:"member"`
	LastTouched int64  `json:"last_touched"`
	Replies     int    `json:"replies"`
}
