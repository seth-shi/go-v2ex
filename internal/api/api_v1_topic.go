package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/alphadose/haxmap"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

const (
	latestNode = "latest"
	latestUri  = "/api/topics/latest.json"
	hotNode    = "hot"
	hotUri     = "/api/topics/hot.json"
)

var (
	// 最新最大
	v1CacheTopics   = haxmap.New[string, response.Topic](3)
	ErrorNoMoreData = errors.New("无更多数据")
)

func (client *v2exClient) getV1Topics(
	ctx context.Context,
	nodeName string,
	page int,
) (*response.Topic, error) {

	var (
		v1Error response.V1ApiError
		v1Res   []response.V1TopicResult
		rr      *resty.Response
		err     error
		uri     = lo.If(nodeName == hotNode, hotUri).Else(latestUri)
	)

	// v1 接口没有分页, 所以我们从缓存中伪造出来
	res, exists := v1CacheTopics.Get(uri)
	if !exists {
		// 请求接口
		rr, err = client.
			client.
			R().
			SetContext(ctx).
			SetResult(&v1Res).
			SetError(&v1Error).
			Get(uri)
		if err != nil {
			return nil, err
		}

		if !v1Error.IsSuccess() {
			return nil, fmt.Errorf("主题[%s]%s", rr.Status(), v1Error.Message)
		}

		topics := lo.Map(
			v1Res, func(item response.V1TopicResult, index int) response.TopicResult {
				return response.TopicResult{
					Id:          item.Id,
					Node:        item.Node.Title,
					Title:       item.Title,
					Member:      item.Member.Username,
					LastTouched: item.LastTouched,
					Replies:     item.Replies,
				}
			},
		)
		res = response.Topic{Items: topics}
		v1CacheTopics.Set(uri, res)
	}

	// 从数据中构建出来
	res.Pagination = response.Page{
		TotalCount: len(res.Items),
		CurrPage:   page,
	}
	res.Pagination.ResetPerPageTo10()
	res.Items = lo.Subset(res.Items, (page-1)*perPage, perPage)
	if len(res.Items) == 0 {
		return nil, ErrorNoMoreData
	}

	return &res, nil
}
