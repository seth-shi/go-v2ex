package api

import (
	"github.com/seth-shi/go-v2ex/v2/api/internal/api_topics"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/pkg"
	"resty.dev/v3"
)

const (
	baseUrl = "https://www.v2ex.com"
)

var (
	V2ex = &v2exClient{}
)

type v2exClient struct {
	client     *resty.Client
	v1TopicApi *api_topics.V1TopicApi
	v2TopicApi *api_topics.V2TopicApi
}

func SetUpHttpClient(conf *model.FileConfig) {

	// 初始化 http 客户端
	client := pkg.NewHTTPClient(conf).
		SetBaseURL(baseUrl).
		AddRequestMiddleware(beforeRequest).
		AddResponseMiddleware(apiErrorHandler).
		AddResponseMiddleware(rateLimitHandler)

	if conf.IsMockEnv() {
		// 默认使用 V2 接口
		g.Config.Update(
			func(conf *model.FileConfig) {
				conf.ChooseAPIV2 = true
			},
		)
		client.SetTransport(&pkg.MockRoundTripper{Mock: mockApiResp})
	} else if conf.IsDevelopmentEnv() {
		client.SetResponseBodyUnlimitedReads(true)
	}

	V2ex = &v2exClient{
		v1TopicApi: api_topics.NewV1(client),
		v2TopicApi: api_topics.NewV2(client),
		client:     client,
	}
}

func (cli *v2exClient) GetLimitRate() float64 {
	return getLimitRate()
}
