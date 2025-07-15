package response

import (
	"strings"
)

type TopicResult struct {
	Id          int64          `json:"id"`
	Title       string         `json:"title"`
	Replies     int            `json:"replies"`
	Member      MemberResult   `json:"member"`
	Node        NodeInfoResult `json:"node"`
	Created     int64          `json:"created"`
	LastTouched int64          `json:"last_touched"`
}

func (t TopicResult) GetTitle() string {
	return strings.ReplaceAll(strings.ReplaceAll(t.Title, "\n", ""), "\r", "")
}
