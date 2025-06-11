package api

import (
	"io"
	"time"

	"resty.dev/v3"

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
}

func newClient() *v2exClient {

	logger := logrus.New()
	logger.Out = io.Discard

	client := &v2exClient{}

	// 初始化 http 客户端
	restyClient := resty.
		New().
		SetBaseURL(baseUrl).
		SetTimeout(time.Second * 10).
		SetLogger(logger).
		AddRequestMiddleware(beforeRequest).
		AddResponseMiddleware(apiErrorHandler).
		AddResponseMiddleware(rateLimitHandler)
	client.client = restyClient

	return client
}

func (cli *v2exClient) GetLimitRate() float64 {
	return getLimitRate()
}
