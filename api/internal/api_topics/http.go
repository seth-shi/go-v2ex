package api_topics

import (
	"context"

	"github.com/seth-shi/go-v2ex/pkg"
	"github.com/seth-shi/go-v2ex/response"
)

func (api *TopicGroupApi) requestTopics(
	ctx context.Context,
	nodeKey string,
) (
	*response.GroupTopic, error,
) {
	var (
		res *response.GroupTopic
		err error
	)
	// 必须这样子获取分页参数
	page, exists := api.nodeRequestPageState.Load(nodeKey)
	if !exists {
		page = 1
	}

	// 请求之前先检查二级是否超过最大分页
	total, exists := api.nodeTotalCount.Load(nodeKey)
	if exists && page > pkg.TotalPages(total, officialPerPage) {
		return nil, ErrNodeApiNoMorePage
	}

	switch nodeKey {
	case latestNode, hotNode:
		res, err = api.requestV1Topics(ctx, nodeKey, page)
	default:
		res, err = api.requestV2Topics(ctx, nodeKey, page)
	}

	// 如果获取成功, 那么缓存数据
	if err == nil {
		// 设置最大页数
		api.nodeRequestPageState.Store(nodeKey, page+1)
	}

	// 不管是 v1 还是 v2 转成成统一的数据格式
	return res, err
}
