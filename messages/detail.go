package messages

import (
	"github.com/seth-shi/go-v2ex/v2/response"
)

type GetDetailResponse struct {
	TopicID   int64
	RequestID uint64
	Data      response.V2DetailResult
}

type GetReplyResponse struct {
	TopicID   int64
	RequestID uint64
	Page      int
	Data      response.V2ReplyResponse
}

type DecodeDetailContentResult struct {
	TopicID   int64
	RequestID uint64
	Result    map[string]string
}

type RenderDetailContentResult struct {
	Content string
}
