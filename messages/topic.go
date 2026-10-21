package messages

import (
	"github.com/seth-shi/go-v2ex/v2/response"
)

type GetTopicResponse struct {
	NodeKey    string
	RequestID  uint64
	Data       []response.TopicResult
	PageInfo   *response.PerTenPageInfo
	CachePages int
}
