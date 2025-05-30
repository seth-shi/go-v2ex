package http

import (
	"io"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/sirupsen/logrus"
)

var (
	V2exClient = newClient()
)

type v2exClient struct {
	client *resty.Client
}

type apiError struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
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

func (client *v2exClient) SetConfig(cfg config.FileConfig) *v2exClient {

	client.client.SetTimeout(time.Second * time.Duration(cfg.Timeout))
	client.client.SetAuthToken(cfg.Token)
	return client
}
