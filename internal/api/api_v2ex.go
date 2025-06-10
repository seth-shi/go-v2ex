package api

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/seth-shi/go-v2ex/internal/config"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const (
	baseUrl = "https://www.v2ex.com"
)

var (
	V2ex = newClient()
)

type v2exClient struct {
	client *resty.Client
	// header 返回的速率限制
	limitRemainCount *atomic.Int64
	limitTotalCount  *atomic.Int64
}

func newClient() *v2exClient {

	logger := logrus.New()
	logger.Out = io.Discard

	client := &v2exClient{
		limitRemainCount: &atomic.Int64{},
		limitTotalCount:  &atomic.Int64{},
	}

	// 初始化 http 客户端
	restyClient := resty.
		New().
		SetBaseURL(baseUrl).
		SetTimeout(time.Second * 10).
		SetLogger(logger).
		OnAfterResponse(client.setRateLimitHandler)
	client.client = restyClient

	return client
}

func (client *v2exClient) RefreshConfig() *v2exClient {

	client.client.SetTimeout(time.Second * time.Duration(config.G.Timeout))
	client.client.SetAuthToken(config.G.Token)
	return client
}
