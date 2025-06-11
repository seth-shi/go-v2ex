package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

const (
	v2TopicsUri = "/api/v2/nodes/%s/topics?p=%d"
	perPage     = 10
)

var (
	// 保证每页都是 10 个
	// v2 只预缓存一页
	v2CacheTopicsRes response.Topic
	v2CacheTopicsKey string
	v2CacheLocker    sync.Mutex
)

func (cli *v2exClient) getV2Topics(ctx context.Context, nodeName string, page int) (*response.Topic, error) {

	// 使用 V2 的接口
	var (
		v2Res response.V2Topic
		res   response.Topic
		// 从缓存中获取 key
		cacheKey = fmt.Sprintf("%s_%d", nodeName, page)
		// 第一页返回: 0~0, 否则返回: 10~20
		retOffset = lo.If(page%2 == 1, 0).Else(perPage)
	)

	// 先从缓存中获取下一页的数据, 如果是偶数页, 则取本页数据, 否则取下一页数据
	v2CacheLocker.Lock()
	if cacheKey == v2CacheTopicsKey {
		// 截断前 10 个放在本页, 后 10 个缓存到下一页
		res = v2CacheTopicsRes
	}
	v2CacheLocker.Unlock()

	// 如果没有缓存, 去接口里请求数据
	if len(res.Items) == 0 {
		// 去请求 API 获取数据, api 分页需要处理一下
		var (
			apiRequestPage = (page + 1) / 2
			requestUri     = fmt.Sprintf(v2TopicsUri, nodeName, apiRequestPage)
		)
		_, err := cli.
			client.
			R().
			SetContext(ctx).
			SetResult(&v2Res).
			Get(requestUri)
		if err != nil {
			return nil, err
		}

		// 预先缓存一页, 由于接口返回 20 个一页, 这边使用切换调整成 10 个一页
		v2Res.Pagination.ResetPerPageTo10()
		res = response.Topic{
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
		}

		v2CacheLocker.Lock()
		defer v2CacheLocker.Unlock()
		v2CacheTopicsKey = requestUri
		v2CacheTopicsRes = res

	}

	res.Pagination.CurrPage = page
	res.Items = lo.Subset(res.Items, retOffset, perPage)
	return &res, nil
}
