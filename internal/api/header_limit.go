package api

import (
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

const (
	headerLimit  = "x-rate-limit-limit"
	headerRemain = "x-rate-limit-remaining"
)

func (client *v2exClient) setRateLimitHandler(c *resty.Client, r *resty.Response) error {

	limit, err := strconv.ParseInt(r.Header().Get(headerLimit), 10, 64)
	if err == nil {
		client.limitTotalCount.Store(limit)
	}
	remain, err := strconv.ParseInt(r.Header().Get(headerRemain), 10, 64)
	if err == nil {
		client.limitRemainCount.Store(remain)
		if remain == 0 {
			return response.ErrTokenLimit
		}
	}

	return nil
}

func (client *v2exClient) GetLimitRate() float64 {

	total := client.limitRemainCount.Load()
	if total <= 0 {
		return 0
	}

	return float64(client.limitRemainCount.Load()) / float64(total)
}
