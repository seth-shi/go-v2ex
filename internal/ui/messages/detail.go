package messages

import "github.com/seth-shi/go-v2ex/internal/types"

type GetDetailRequest struct {
	ID int64
}

type GetDetailResult struct {
	Detail types.V2DetailResult
}
