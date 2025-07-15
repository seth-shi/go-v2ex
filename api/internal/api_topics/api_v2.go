package api_topics

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync/atomic"

	"github.com/puzpuzpuz/xsync/v4"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/g"
	"github.com/seth-shi/go-v2ex/response"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

const (
	v2TopicsUri = "/api/v2/nodes/%s/topics?p=%d"
)

// V2TopicApi v1 和 v2 完全隔离开, 防止后续影响
type V2TopicApi struct {
	client *resty.Client
	// 内部记录节点请求第几页的状态
	groupTotalCount   *xsync.Map[string, int]
	nodePageInfoState *xsync.Map[string, response.V2PageResponse]
	cursorPageState   *xsync.Map[string, int]
	cacheData         []response.TopicResult
	cacheNode         string
	// 只缓存某一个节点的数据, 当节点切换, 立即清空数据
	isRequesting atomic.Bool
}

type v2Resp struct {
	Total  int
	Result []response.TopicResult
}

func NewV2(client *resty.Client) *V2TopicApi {
	return &V2TopicApi{
		client:            client,
		groupTotalCount:   xsync.NewMap[string, int](),
		nodePageInfoState: xsync.NewMap[string, response.V2PageResponse](),
		cursorPageState:   xsync.NewMap[string, int](),
	}
}

func (api *V2TopicApi) GetTopicsByGroupNode(
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
	endIndex := (page-1)*perPage + perPage - 1
	if len(api.cacheData) < endIndex {
		// 如果没有缓存, 或者当前页码数据不在缓存中, 那么去从接口聚合获取
		var result []response.TopicResult
		result, err = api.groupRequestData(node)
		if err != nil {
			return
		}

		// 组装数据存到缓存中
		api.cacheData = append(api.cacheData, result...)
	}

	res = lo.Subset(api.cacheData, (page-1)*perPage, perPage)
	total, _ = api.groupTotalCount.Load(node.Key)
	return res, total, nil
}

func (api *V2TopicApi) prepareRequest(node g.GroupNode, page int) error {

	// 如果节点不是上次的, 清空缓存
	if node.Key != api.cacheNode {
		api.cacheData = nil
		api.cacheNode = node.Key
	}

	return nil
}

func (api *V2TopicApi) finishRequest(err *error) {
	api.isRequesting.Store(false)
	if r := recover(); r != nil {
		// 这里的 *err 赋值是安全的
		*err = fmt.Errorf("%+v", r)
		slog.Info("请求主题失败", "err", *err)
	}
}

func (api *V2TopicApi) groupRequestData(groupNode g.GroupNode) (
	[]response.TopicResult,
	error,
) {
	// 并发获取子节点的数据
	var (
		ctx      = context.Background()
		chResult = make(chan *v2Resp)
		chError  = make(chan error, 1)
		eg       errgroup.Group
	)
	slog.Info(
		"v2正在请求主题",
		slog.String("节点名", groupNode.Name),
		slog.Int("节点数", len(groupNode.Nodes)),
	)
	for _, nodeKey := range groupNode.Nodes {
		// 错误统一返回
		eg.Go(
			func() error {
				resp, err := api.requestV2Topics(ctx, nodeKey)
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
		total   int
		results []response.TopicResult
	)
	for res := range chResult {
		total += res.Total
		results = append(results, res.Result...)
		slog.Info(
			"v2单次数据返回",
			slog.Int("v2数据量", len(res.Result)),
		)
	}
	// 等待所有任务完成, 并返回可能出错的结果
	if err := <-chError; err != nil {
		return nil, err
	}

	// 如果已经有了 key, 那么就不存储, 防止分页到后面无法获取数据
	api.groupTotalCount.LoadOrStore(groupNode.Key, total)
	// 如果不是最新/最热, 那么按照时间重新排序依次
	slices.SortFunc(
		results, func(a, b response.TopicResult) int {
			return cmp.Compare(b.LastTouched, a.LastTouched)
		},
	)

	return results, nil
}

func (api *V2TopicApi) requestV2Topics(
	ctx context.Context,
	nodeKey string,
) (
	*v2Resp,
	error,
) {

	// 请求下一页
	apiPage, _ := api.cursorPageState.LoadOrStore(nodeKey, 0)
	apiPage++

	// 把官方分页参数存储起来
	pi, exists := api.nodePageInfoState.Load(nodeKey)
	// 官方分页是 20 个一条
	slog.Info("v2API请求", slog.String("node", nodeKey), slog.Int("page", apiPage))
	if exists && apiPage > pi.TotalPages {
		return nil, ErrNodeApiNoMorePage
	}

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

	// 如果返回了最大页数, 那么返回给前端, 方便后续缓存获取
	api.cursorPageState.Store(nodeKey, apiPage)
	// 存储分页信息
	api.nodePageInfoState.Store(nodeKey, v2Res.Pagination)

	res := &v2Resp{
		Total: v2Res.Pagination.TotalCount,
		Result: lo.Map(
			v2Res.Result, func(item response.V2TopicResult, index int) response.TopicResult {
				return response.TopicResult{
					Id: item.Id,
					Node: response.NodeInfoResult{
						Id:    0,
						Name:  nodeKey,
						Title: nodeKey,
					},
					Title: item.Title,
					Member: response.MemberResult{
						Username: item.LastReplyBy,
					},
					LastTouched: item.LastTouched,
					Replies:     item.Replies,
				}
			},
		),
	}
	return res, nil
}
