package messages

import (
	"github.com/seth-shi/go-v2ex/internal/types"
)

type GetTopicsRequest struct {
	Page      int
	NodeIndex int
}

type GetTopicsResult struct {
	Topics []types.TopicComResult
	Page   int
	Error  error
}
