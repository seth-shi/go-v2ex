package messages

import (
	"github.com/seth-shi/go-v2ex/internal/types"
)

type GetTopicsRequest struct {
	Page int
}

type GetTopicsResult struct {
	Topics     []types.TopicComResult
	Pagination types.Pagination
	// 监听者需要处理请求回调 (请求拦截)
	Error error
}
