package api_topics

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

const (
	v2TopicsUri = "/api/v2/nodes/%s/topics?p=%d"
)

func (api *TopicGroupApi) requestV2Topics(ctx context.Context, nodeKey string, apiPage int) (
	*response.GroupTopic, error,
) {
	// 去请求 API 获取数据, api 分页需要处理一下
	var (
		v2Res      response.V2TopicResponse
		requestUri = fmt.Sprintf(v2TopicsUri, nodeKey, apiPage)
	)

	_, err := api.
		client.
		R().
		SetContext(ctx).
		SetResult(&v2Res).
		Get(requestUri)
	if err != nil {
		return nil, err
	}

	// !!! 最新最热, 永远只能有一页
	if v2Res.Pagination.TotalCount > 0 {
		api.nodeTotalCount.Store(nodeKey, v2Res.Pagination.TotalCount)
	}

	pageInfo := &response.PerTenPageInfo{
		TotalCount: v2Res.Pagination.TotalCount,
		CurrPage:   apiPage,
	}
	return &response.GroupTopic{
		Items: lo.Map(
			v2Res.Result, func(item response.V2TopicResult, index int) response.TopicResult {
				return response.TopicResult{
					Id:          item.Id,
					Node:        nodeKey,
					Title:       item.GetTitle(),
					Member:      item.LastReplyBy,
					LastTouched: item.LastTouched,
					Replies:     item.Replies,
				}
			},
		),
		Pagination: pageInfo,
	}, nil
}
