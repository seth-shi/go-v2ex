package api

import (
	"io"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/seth-shi/go-v2ex/internal/config"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

var (
	Client           = newClient()
	LimitRemainCount = &atomic.Int64{}
	LimitTotalCount  = &atomic.Int64{}
)

type v2exClient struct {
	client *resty.Client
}

func newClient() *v2exClient {

	logger := logrus.New()
	logger.Out = io.Discard
	return &v2exClient{
		client: resty.
			New().
			SetBaseURL("https://www.v2ex.com").
			SetTimeout(time.Second * 10).
			OnAfterResponse(func(c *resty.Client, r *resty.Response) error {

				limit, err := strconv.ParseInt(r.Header().Get("x-rate-limit-limit"), 10, 64)
				if err == nil {
					LimitTotalCount.Store(limit)
				}
				remain, err := strconv.ParseInt(r.Header().Get("x-rate-limit-remaining"), 10, 64)
				if err == nil {
					LimitRemainCount.Store(remain)
				}

				return nil
			}).
			SetLogger(logger),
	}
}

func (client *v2exClient) RefreshConfig() *v2exClient {

	client.client.SetTimeout(time.Second * time.Duration(config.G.Timeout))
	client.client.SetAuthToken(config.G.Token)
	return client
}
