package messages

import (
	"github.com/seth-shi/go-v2ex/v2/response"
)

type GetDetailResponse struct {
	Data response.V2DetailResult
}

type GetReplyResponse struct {
	Data response.V2ReplyResponse
}

type DecodeDetailContentResult struct {
	Result map[string]string
}

type RenderDetailContentResult struct {
	Content string
}
