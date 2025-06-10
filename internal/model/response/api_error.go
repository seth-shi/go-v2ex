package response

import (
	"errors"
)

var (
	ErrNoMoreData     = errors.New("无更多数据")
	ErrTokenLimit     = errors.New("令牌已被限制请求,请稍后再试")
	ErrRequestTimeout = errors.New("请求超时,请确保已配置代理")
)

type V2ApiError struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type V1ApiError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (e V1ApiError) IsSuccess() bool {
	return e.Status == ""
}

func (e V2ApiError) IsSuccess() bool {
	return e.Success
}
