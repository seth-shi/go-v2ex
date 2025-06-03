package messages

import "github.com/seth-shi/go-v2ex/internal/types"

type GetRepliesResult struct {
	Replies    []types.V2ReplyResult
	Pagination types.Pagination
	Error      error
}
