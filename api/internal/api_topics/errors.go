package api_topics

import (
	"errors"
)

var (
	ErrNodeApiNoMorePage  = errors.New("没有分页数据")
	ErrNoMoreData         = errors.New("无更多数据")
	ErrLockingRequestData = errors.New("正在请求数据")
)
