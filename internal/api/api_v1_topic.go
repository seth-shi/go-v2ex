package api

import (
	"context"

	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/model/response"
	"github.com/seth-shi/go-v2ex/internal/pkg"
)

const (
	latestNode = "latest"
	latestUri  = "/api/topics/latest.json"
	hotNode    = "hot"
	hotUri     = "/api/topics/hot.json"
)

var (
	// 最新最大
	v1TopicCache = pkg.NewLockableMap[response.Topic](10)
)

func (cli *v2exClient) getV1Topics(
	ctx context.Context,
	nodeName string,
	page int,
) (*response.Topic, error) {

	var (
		res    response.Topic
		exists bool
		err    error
	)

	// v1 接口没有分页, 所以我们从缓存中伪造出来
	if res, exists = v1TopicCache.Get(nodeName); !exists {
		if res, err = cli.requestV1Topics(ctx, nodeName); err != nil {
			return nil, err
		}
		// 写入缓存
		v1TopicCache.Set(nodeName, res)
	}

	// 从数据中构建出来
	res.Pagination = response.Page{
		TotalCount: len(res.Items),
		CurrPage:   page,
	}
	res.Pagination.ResetPerPageTo10()
	res.Items = lo.Subset(res.Items, (page-1)*perPage, perPage)
	if len(res.Items) == 0 {
		return nil, ErrNoMoreData
	}

	return &res, nil
}

func (cli *v2exClient) requestV1Topics(ctx context.Context, nodeName string) (response.Topic, error) {

	var (
		err   error
		v1Res []response.V1TopicResult
		uri   = lo.If(nodeName == hotNode, hotUri).Else(latestUri)
	)
	// 去请求 API 获取数据, api 分页需要处理一下
	_, err = cli.
		client.
		R().
		SetContext(ctx).
		SetResult(&v1Res).
		Get(uri)
	if err != nil {
		return response.Topic{}, err
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
	return response.Topic{Items: topics}, nil
}
