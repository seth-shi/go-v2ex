package response

import (
	"strings"
)

type V2TopicResponse struct {
	Result     []V2TopicResult `json:"result"`
	Pagination V2PageResponse  `json:"pagination"`
}

type V2TopicResult struct {
	Id           int64  `json:"id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Replies      int    `json:"replies"`
	LastReplyBy  string `json:"last_reply_by"`
	Created      int    `json:"created"`
	LastModified int64  `json:"last_modified"`
	LastTouched  int64  `json:"last_touched"`
}

func (t V2TopicResult) GetTitle() string {
	return strings.ReplaceAll(strings.ReplaceAll(t.Title, "\n", ""), "\r", "")
}
