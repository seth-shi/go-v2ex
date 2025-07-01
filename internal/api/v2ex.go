package api

import (
	"github.com/seth-shi/go-v2ex/internal/api/api_topics"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/pkg"
	"resty.dev/v3"
)

const (
	baseUrl = "https://www.v2ex.com"
)

var (
	V2ex = &v2exClient{}
)

type v2exClient struct {
	client   *resty.Client
	topicApi *api_topics.TopicGroupApi
}

func SetUpHttpClient(conf *config.FileConfig) {

	// 初始化 http 客户端
	client := pkg.NewHTTPClient(conf).
		SetBaseURL(baseUrl).
		AddRequestMiddleware(beforeRequest).
		AddResponseMiddleware(apiErrorHandler).
		AddResponseMiddleware(rateLimitHandler)

	if conf.IsMockEnv() {
		client.SetTransport(&pkg.MockRoundTripper{Mock: mockApiResp})
	}

	V2ex = &v2exClient{
		topicApi: api_topics.New(client),
		client:   client,
	}
}

func (cli *v2exClient) GetLimitRate() float64 {
	return getLimitRate()
}
