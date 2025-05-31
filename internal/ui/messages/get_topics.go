package messages

import (
	"github.com/seth-shi/go-v2ex/internal/types"
)

type GetTopicsRequest struct {
	Page int
}

type GetTopicsResult struct {
	Topics []*types.TopicResource
	Page   int
	Error  error
}
