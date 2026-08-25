package response

import (
	"strings"
)

type TopicResult struct {
	Id           int64          `json:"id"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	Url          string         `json:"url"`
	Replies      int            `json:"replies"`
	Visits       int            `json:"visits"`
	Member       MemberResult   `json:"member"`
	LastReplyBy  string         `json:"last_reply_by,omitempty"`
	Node         NodeInfoResult `json:"node"`
	Created      int64          `json:"created"`
	LastModified int64          `json:"last_modified"`
	LastTouched  int64          `json:"last_touched"`
}

func (t TopicResult) GetTitle() string {
	return strings.ReplaceAll(strings.ReplaceAll(t.Title, "\n", ""), "\r", "")
}
