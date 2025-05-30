package api

import (
	"io"
	"time"

	"github.com/seth-shi/go-v2ex/internal/types"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

var (
	Client = newClient()
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
			SetLogger(logger),
	}
}

func (client *v2exClient) SetConfig(cfg types.FileConfig) *v2exClient {

	client.client.SetTimeout(time.Second * time.Duration(cfg.Timeout))
	client.client.SetAuthToken(cfg.Token)
	return client
}
