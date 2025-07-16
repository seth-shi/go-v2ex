package api_topics

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync/atomic"

	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/response"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

const (
	perPage = 10
)

const (
	latestUri = "/api/topics/latest.json"
	hotUri    = "/api/topics/hot.json"
	otherUri  = "/api/topics/show.json?node_name=%s"
)

type V1TopicApi struct {
	client *resty.Client
	// 内部记录节点请求第几页的状态
	cacheData []response.TopicResult
	cacheNode string
	// 只缓存某一个节点的数据, 当节点切换, 立即清空数据
	isRequesting atomic.Bool
}

func NewV1(client *resty.Client) *V1TopicApi {
	return &V1TopicApi{
		client: client,
	}
}

func (api *V1TopicApi) GetTopicsByGroupNode(
	ctx context.Context,
	node g.GroupNode,
	page int,
) (res []response.TopicResult, total int, err error) {

	// 只允许单个请求进来获取 数据
	if !api.isRequesting.CompareAndSwap(false, true) {
		return nil, 0, ErrLockingRequestData
	}
	defer api.finishRequest(&err)

	// 请求前置处理, 清空缓存等等
	if err = api.prepareRequest(node, page); err != nil {
		return
	}

	// 如果有存储了最大条数, 并且当前页码超过, 那么直接返回无数据
	if len(api.cacheData) == 0 {
		// 如果没有缓存, 或者当前页码数据不在缓存中, 那么去从接口聚合获取
		var result []response.TopicResult
		result, err = api.groupRequestData(node)
		if err != nil {
			return
		}

		// 组装数据存到缓存中
		api.cacheData = result
	}

	res = lo.Subset(api.cacheData, (page-1)*perPage, perPage)
	total = len(api.cacheData)
	return res, total, nil
}

func (api *V1TopicApi) prepareRequest(node g.GroupNode, page int) error {

	// 如果节点不是上次的, 清空缓存
	if node.Key != api.cacheNode {
		api.cacheData = nil
		api.cacheNode = node.Key
	}

	return nil
}

func (api *V1TopicApi) finishRequest(err *error) {
	api.isRequesting.Store(false)
	if r := recover(); r != nil {
		// 这里的 *err 赋值是安全的
		*err = fmt.Errorf("%+v", r)
		slog.Info("请求主题失败", "err", *err)
	}
}

func (api *V1TopicApi) groupRequestData(groupNode g.GroupNode) (
	[]response.TopicResult,
	error,
) {
	// 并发获取子节点的数据
	var (
		ctx      = context.Background()
		chResult = make(chan []response.TopicResult)
		chError  = make(chan error, 1)
		eg       errgroup.Group
	)
	slog.Info(
		"正在请求主题",
		slog.String("节点名", groupNode.Name),
		slog.Int("节点数", len(groupNode.Nodes)),
	)
	for _, nodeKey := range groupNode.Nodes {
		// 错误统一返回
		eg.Go(
			func() error {
				resp, err := api.requestV1Topics(ctx, nodeKey)
				if err == nil {
					chResult <- resp
				}

				// 如果错误是么有更多数据, 那么不返回
				if errors.Is(err, ErrNodeApiNoMorePage) {
					return nil
				}

				return err
			},
		)
	}
	go func() {
		chError <- eg.Wait()
		close(chResult)
	}()

	// 数据合并到一个数组里, 分页参数累加
	var (
		results []response.TopicResult
	)
	for res := range chResult {
		results = append(results, res...)
		slog.Info(
			"单次数据返回",
			slog.Int("数据量", len(res)),
		)
	}
	// 等待所有任务完成, 并返回可能出错的结果
	if err := <-chError; err != nil {
		return nil, err
	}

	// 如果不是最新/最热, 那么按照时间重新排序依次
	if groupNode.Key != g.HotNode && groupNode.Key != g.LatestNode {
		slices.SortFunc(
			results, func(a, b response.TopicResult) int {
				return cmp.Compare(b.LastTouched, a.LastTouched)
			},
		)
	}

	return results, nil
}

func (api *V1TopicApi) requestV1Topics(
	ctx context.Context,
	nodeKey string,
) (
	[]response.TopicResult, error,
) {

	var (
		err   error
		v1Res []response.TopicResult
		uri   = lo.
			If(nodeKey == g.HotNode, hotUri).
			ElseIf(nodeKey == g.LatestNode, latestUri).
			Else(fmt.Sprintf(otherUri, nodeKey))
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

	return v1Res, nil
}
