package api_topics

import (
	"context"

	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

const (
	latestNode = "latest"
	latestUri  = "/api/topics/latest.json"
	hotNode    = "hot"
	hotUri     = "/api/topics/hot.json"
)

func (api *TopicGroupApi) requestV1Topics(
	ctx context.Context, nodeKey string, page int,
) (*response.GroupTopic, error) {

	if page > 1 {
		return nil, ErrNodeApiNoMorePage
	}

	var (
		err   error
		v1Res []response.V1TopicResult
		uri   = lo.If(nodeKey == hotNode, hotUri).Else(latestUri)
	)
	// 去请求 API 获取数据, api 分页需要处理一下
	_, err = api.
		client.
		R().
		SetContext(ctx).
		SetResult(&v1Res).
		Get(uri)
	if err != nil {
		return nil, err
	}

	topics := lo.Map(
		v1Res, func(item response.V1TopicResult, index int) response.TopicResult {
			return response.TopicResult{
				Id:          item.Id,
				Node:        item.Node.Title,
				Title:       item.GetTitle(),
				Member:      item.Member.Username,
				LastTouched: item.LastTouched,
				Replies:     item.Replies,
			}
		},
	)

	api.nodeTotalCount.Store(nodeKey, len(topics))
	resPage := &response.PerTenPageInfo{
		TotalCount: len(topics),
		CurrPage:   page,
	}
	return &response.GroupTopic{Items: topics, Pagination: resPage}, nil
}
