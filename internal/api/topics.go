package api

import (
	"context"
	"errors"
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

const (
	latestNode  = "latest"
	latestUri   = "/api/topics/latest.json"
	hotNode     = "hot"
	hotUri      = "/api/topics/hot.json"
	v2TopicsUri = "/api/v2/nodes/%s/topics?p=%d"
	perPage     = 10
)

var (
	// 保证每页都是 10 个
	v1CacheTopics = make(map[string][]types.TopicComResult)
	v1CacheLocker sync.Mutex
	// v2 只预缓存一页
	v2CacheTopics    []types.TopicComResult
	v2CacheTopicsKey string
	v2CacheLocker    sync.Mutex
)

func (client *v2exClient) GetTopics(nodeIndex int, page int) tea.Cmd {

	return func() tea.Msg {

		if page <= 0 {
			page = 1
		}

		var (
			nodeName = lo.NthOr(config.G.GetNodes(), nodeIndex, latestNode)
			res      []types.TopicComResult
			err      error
		)

		// 请求的时候, 用数据的分页数据
		switch nodeName {
		case latestNode, hotNode:
			res, err = client.getV1Topics(nodeName, page)
		default:
			res, err = client.getV2Topics(nodeName, page)
		}

		if err != nil {
			return messages.GetTopicsResult{Error: err}
		}

		return messages.GetTopicsResult{
			Topics: res,
			Page:   page,
		}
	}
}

func (client *v2exClient) getV2Topics(nodeName string, page int) ([]types.TopicComResult, error) {

	// 使用 V2 的接口
	var (
		v2Res types.V2TopicResponse
		rr    *resty.Response
		err   error
		// 从缓存中获取 key
		cacheKey = fmt.Sprintf("%s_%d", nodeName, page)
		cacheRes []types.TopicComResult
		// 第一页返回: 0~0, 否则返回: 10~20
		retOffset = lo.If(page%2 == 1, 0).Else(perPage)
	)

	// 先从缓存中获取下一页的数据, 如果是偶数页, 则取本页数据, 否则取下一页数据
	v2CacheLocker.Lock()
	if cacheKey == v2CacheTopicsKey {
		// 截断前 10 个放在本页, 后 10 个缓存到下一页
		cacheRes = lo.Subset(v2CacheTopics, retOffset, perPage)
	}
	v2CacheLocker.Unlock()
	if len(cacheRes) > 0 {
		return cacheRes, nil
	}

	// 去请求 API 获取数据, api 分页需要处理一下
	var (
		apiRequestPage = (page + 1) / 2
		requestUri     = fmt.Sprintf(v2TopicsUri, nodeName, apiRequestPage)
	)
	rr, err = client.
		client.
		R().
		SetContext(context.Background()).
		SetResult(&v2Res).
		SetError(&v2Res).
		Get(requestUri)
	if err != nil {
		return nil, err
	}

	if !v2Res.Success {
		return nil, fmt.Errorf("[%s]%s", rr.Status(), v2Res.Message)
	}

	res := lo.Map(
		v2Res.Result, func(item types.V2TopicResult, index int) types.TopicComResult {
			return types.TopicComResult{
				Id:          item.Id,
				Node:        nodeName,
				Title:       item.Title,
				Member:      item.LastReplyBy,
				LastTouched: item.LastTouched,
				Replies:     item.Replies,
			}
		},
	)
	// 预先缓存一页, 由于接口返回 20 个一页, 这边使用切换调整成 10 个一页
	v2CacheLocker.Lock()
	defer v2CacheLocker.Unlock()
	v2CacheTopicsKey = requestUri
	res = lo.Subset(res, retOffset, perPage)
	return res, nil
}

func (client *v2exClient) getV1Topics(nodeName string, page int) ([]types.TopicComResult, error) {

	var (
		v1Error types.V1ApiError
		v1Res   []types.V1TopicResult
		rr      *resty.Response
		err     error
		uri     = lo.If(nodeName == hotNode, hotUri).Else(latestUri)
	)

	// 大于第一页的, 只能从缓存中获取
	v1CacheLocker.Lock()
	res, exists := v1CacheTopics[uri]
	v1CacheLocker.Unlock()
	if exists {
		res = lo.Subset(res, (page-1)*perPage, perPage)
		if len(res) > 0 {
			return res, nil
		}

		return nil, errors.New("无更多数据")
	}

	rr, err = client.
		client.
		R().
		SetContext(context.Background()).
		SetResult(&v1Res).
		SetError(&v1Error).
		Get(uri)
	if err != nil {
		return nil, err
	}

	if !v1Error.Success() {
		return nil, fmt.Errorf("[%s]%s", rr.Status(), v1Error.Message)
	}

	res = lo.Map(
		v1Res, func(item types.V1TopicResult, index int) types.TopicComResult {
			return types.TopicComResult{
				Id:          item.Id,
				Node:        item.Node.Title,
				Title:       item.Title,
				Member:      item.Member.Username,
				LastTouched: item.LastTouched,
				Replies:     item.Replies,
			}
		},
	)

	// 开始处理缓存的数据
	v1CacheLocker.Lock()
	defer v1CacheLocker.Unlock()
	v1CacheTopics[uri] = res

	return lo.Subset(res, (page-1)*perPage, perPage), nil
}
