package pkg

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/seth-shi/go-v2ex/v2/model"
	"resty.dev/v3"
)

type ClientOption func(*resty.Client, *model.FileConfig)

func NewHTTPClient(conf *model.FileConfig, options ...ClientOption) *resty.Client {
	client := resty.New()
	client.SetTimeout(time.Second * time.Duration(conf.Timeout))
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	client.SetLogger(RestyLogger())

	// 应用所有配置选项
	for _, opt := range options {
		opt(client, conf)
	}

	return client
}

type MockRoundTripper struct {
	http.RoundTripper
	Mock func(req *http.Request, resp *http.Response)
}

func (dr MockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	slog.Info("mock_request_url", slog.String("url", r.URL.String()))
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	if dr.Mock != nil {
		dr.Mock(r, resp)
	}

	if resp.Body == nil {
		resp.Body = http.NoBody
		resp.ContentLength = 0
	}

	return resp, nil
}
