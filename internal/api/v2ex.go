package api

import (
	"time"

	"github.com/seth-shi/go-v2ex/internal/api/api_topics"
	"github.com/seth-shi/go-v2ex/internal/pkg"
	"resty.dev/v3"
)

const (
	baseUrl = "https://www.v2ex.com"
)

var (
	V2ex = newClient()
)

type v2exClient struct {
	client   *resty.Client
	topicApi *api_topics.TopicGroupApi
}

func newClient() *v2exClient {

	// 初始化 http 客户端
	restyClient := resty.
		New().
		SetBaseURL(baseUrl).
		SetTimeout(time.Second * 10).
		SetLogger(pkg.RestyLogger()).
		AddRequestMiddleware(beforeRequest).
		AddResponseMiddleware(apiErrorHandler).
		AddResponseMiddleware(rateLimitHandler)

	return &v2exClient{
		topicApi: api_topics.New(restyClient),
		client:   restyClient,
	}
}

func (cli *v2exClient) GetLimitRate() float64 {
	return getLimitRate()
}
