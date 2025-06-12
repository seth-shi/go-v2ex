package api

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/model/response"
	"github.com/seth-shi/go-v2ex/internal/pkg"
)

const (
	v2TopicsUri = "/api/v2/nodes/%s/topics?p=%d"
	perPage     = 10
)

var (
	// 不缓存指针, 防止修改
	v2TopicCache = pkg.NewLockableMap[response.Topic](1000)
)

func (cli *v2exClient) getV2Topics(ctx context.Context, nodeName string, page int) (*response.Topic, error) {

	// 使用 V2 的接口
	var (
		res    response.Topic
		exists bool
		err    error
		// 请求的接口页是 1~2=1  3~4=2
		apiPage  = (page + 1) / 2
		cacheKey = fmt.Sprintf("%s_%d", nodeName, apiPage)
		// 第一页返回: 0~10, 否则返回: 10~20
		retOffset = lo.If(page%2 == 1, 0).Else(perPage)
	)

	// 没有缓存的话, 直接从接口获取数据
	if res, exists = v2TopicCache.Get(cacheKey); !exists {
		// 无法从接口获取数据, 直接返回错误
		if res, err = cli.requestV2Topics(ctx, nodeName, apiPage); err != nil {
			return nil, err
		}
		// 写入缓存
		v2TopicCache.Set(cacheKey, res)
	}

	res.Pagination.CurrPage = page
	res.Items = lo.Subset(res.Items, retOffset, perPage)
	return &res, nil
}

func (cli *v2exClient) requestV2Topics(ctx context.Context, nodeName string, apiPage int) (response.Topic, error) {
	// 去请求 API 获取数据, api 分页需要处理一下
	var (
		v2Res      response.V2Topic
		requestUri = fmt.Sprintf(v2TopicsUri, nodeName, apiPage)
	)
	_, err := cli.
		client.
		R().
		SetContext(ctx).
		SetResult(&v2Res).
		Get(requestUri)
	if err != nil {
		return response.Topic{}, err
	}

	// 预先缓存一页, 由于接口返回 20 个一页, 这边使用切换调整成 10 个一页
	v2Res.Pagination.ResetPerPageTo10()
	return response.Topic{
		Items: lo.Map(
			v2Res.Result, func(item response.V2TopicResult, index int) response.TopicResult {
				return response.TopicResult{
					Id:          item.Id,
					Node:        nodeName,
					Title:       item.Title,
					Member:      item.LastReplyBy,
					LastTouched: item.LastTouched,
					Replies:     item.Replies,
				}
			},
		),
		Pagination: v2Res.Pagination,
	}, nil
}
