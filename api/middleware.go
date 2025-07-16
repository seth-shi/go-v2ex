package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/response"
	"resty.dev/v3"
)

const (
	headerLimit          = "x-rate-limit-limit"
	headerRemain         = "x-rate-limit-remaining"
	headerKeyContentType = "Content-Type"
	jsonType             = "application/json"
	apiV2BasePath        = "/api/v2"
)

var (
	limitRemainCount atomic.Int64
	limitTotalCount  atomic.Int64
)

func beforeRequest(client *resty.Client, request *resty.Request) error {
	// 修改后及时生效
	request.SetAuthToken(g.Config.Get().Token)
	return nil
}

func apiErrorHandler(client *resty.Client, resp *resty.Response) error {

	// resp.Err 上一个中间件传递的错误
	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	}

	// 如果不是 JSON, 那么直接返回状态码, 防止 HTML 污染
	if strings.Contains(resp.Header().Get(headerKeyContentType), jsonType) {
		// 序列化成 v1, v2 错误
		var (
			v1Err response.V1ApiError
			v2Err response.V2ApiError
			err   error
		)

		// 只有序列化成功才去返回错误, 否则返回返回默认错误
		if strings.Contains(resp.Request.URL, apiV2BasePath) {
			if err = json.Unmarshal(resp.Bytes(), &v2Err); err == nil {
				return v2Err
			}
		} else {
			if err = json.Unmarshal(resp.Bytes(), &v1Err); err == nil {
				return v1Err
			}
		}
	}

	// 返回默认的错误
	return fmt.Errorf("http[%s]", resp.Status())
}

func rateLimitHandler(c *resty.Client, r *resty.Response) error {

	limit, err := strconv.ParseInt(r.Header().Get(headerLimit), 10, 64)
	if err == nil {
		limitTotalCount.Store(limit)
	}
	remain, err := strconv.ParseInt(r.Header().Get(headerRemain), 10, 64)
	if err == nil {
		limitRemainCount.Store(remain)
	}

	return nil
}

func getLimitRate() float64 {

	total := limitTotalCount.Load()
	if total <= 0 {
		return 0
	}

	return float64(limitRemainCount.Load()) / float64(total)
}
