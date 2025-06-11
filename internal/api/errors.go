package api

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNoMoreData      = errors.New("无更多数据")
	ErrUnauthorized    = errors.New("未配置令牌")
	ErrRequestTimeout  = errors.New("请求超时,请确保已配置代理")
	ErrRequestCanceled = errors.New("请求取消")
)

func errorWrapper(prefix string, err error) error {

	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		err = ErrRequestTimeout
	case errors.Is(err, context.Canceled):
		err = ErrRequestCanceled
	}

	return fmt.Errorf("请求%s错误:%s", prefix, err)
}
