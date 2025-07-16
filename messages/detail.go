package messages

import (
	"github.com/seth-shi/go-v2ex/v2/response"
)

type GetDetailRequest struct {
	ID int64
}

type GetDetailResponse struct {
	Data response.V2DetailResult
}

type GetReplyResponse struct {
	Data     response.V2ReplyResponse
	CurrPage int
}

type GetImageRequest struct {
	URL []string
}

type GetImageResult struct {
	Result map[string]string
}
