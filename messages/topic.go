package messages

import (
	"github.com/seth-shi/go-v2ex/v2/response"
)

type GetTopicResponse struct {
	Data       []response.TopicResult
	PageInfo   *response.PerTenPageInfo
	CachePages int
}
