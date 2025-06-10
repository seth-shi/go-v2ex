package messages

import (
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

type GetDetailRequest struct {
	ID int64
}

type GetDetailResponse struct {
	Data response.V2DetailResult
}

type GetReplyResponse struct {
	Data response.V2Reply
}
