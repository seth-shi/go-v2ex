package types

type V2TopicResponse struct {
	V2ApiError
	Result []V2TopicResult `json:"result"`
}
type V2TopicResult struct {
	Id              int64  `json:"id"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	ContentRendered string `json:"content_rendered"`
	Syntax          int    `json:"syntax"`
	Url             string `json:"url"`
	Replies         int    `json:"replies"`
	LastReplyBy     string `json:"last_reply_by"`
	Created         int    `json:"created"`
	LastModified    int64  `json:"last_modified"`
	LastTouched     int64  `json:"last_touched"`
}

type TopicComResult struct {
	Id          int64  `json:"id"`
	Node        string `json:"node"`
	Title       string `json:"title"`
	Member      string `json:"member"`
	LastTouched int64  `json:"last_touched"`
	Replies     int    `json:"replies"`
}
