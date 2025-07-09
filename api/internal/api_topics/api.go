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
	"github.com/seth-shi/go-v2ex/pkg"
	"github.com/seth-shi/go-v2ex/response"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

const (
	perPage         = 10
	officialPerPage = 20
)

type TopicGroupApi struct {
	client *resty.Client
	// 二级节点状态, 需要一直缓存, 每个节点的
	// 最大页码, 当前请求的页码
	groupTotalCount *xsync.Map[string, int]
	nodeTotalCount  *xsync.Map[string, int]
	// 内部记录节点请求第几页的状态
	nodeRequestPageState *xsync.Map[string, int]
	cacheData            []response.TopicResult
	cacheNode            string
	// 只缓存某一个节点的数据, 当节点切换, 立即清空数据
	isRequesting atomic.Bool
}

func New(client *resty.Client) *TopicGroupApi {
	return &TopicGroupApi{
		client:               client,
		groupTotalCount:      xsync.NewMap[string, int](),
		nodeTotalCount:       xsync.NewMap[string, int](),
		nodeRequestPageState: xsync.NewMap[string, int](),
	}
}

func (api *TopicGroupApi) GetTopicsByGroupNode(
	ctx context.Context,
	node g.GroupNode,
	page int,
) (res *response.GroupTopic, err error) {

	// 只允许单个请求进来获取 数据
	if !api.isRequesting.CompareAndSwap(false, true) {
		return nil, ErrLockingRequestData
	}
	defer api.finishRequest(&err)

	// 请求前置处理, 清空缓存等等
	if err = api.prepareRequest(node, page); err != nil {
		return
	}

	// 如果有存储了最大条数, 并且当前页码超过, 那么直接返回无数据
	var (
		total, exists = api.groupTotalCount.Load(node.Key)
	)
	if exists && page > pkg.TotalPages(total, perPage) {
		return nil, ErrNoMoreData
	}

	// 如果没有缓存, 或者缓存中的数据不足返回到当前页数据, 去获取数据
	if len(api.cacheData) < (page * perPage) {

		// 如果没有缓存, 或者当前页码数据不在缓存中, 那么去从接口聚合获取
		results, pageInfo, err := api.groupRequestData(node)
		if err != nil {
			return nil, err
		}

		// 组装数据存到缓存中
		api.cacheData = append(api.cacheData, results...)
		// 如果已经有数据, 那么就用之前的, 防止情况: A,B,C 子节点翻页之后: A 节点无数据不返回总数
		total, _ = api.groupTotalCount.LoadOrStore(node.Key, pageInfo.TotalCount)
	}

	return api.makeResult(total, page), nil
}

func (api *TopicGroupApi) prepareRequest(node g.GroupNode, page int) error {

	// 如果节点不是上次的, 清空缓存
	if node.Key != api.cacheNode {
		api.cacheData = nil
		api.cacheNode = node.Key
		api.nodeRequestPageState.Clear()
	}

	return nil
}

func (api *TopicGroupApi) finishRequest(err *error) {
	api.isRequesting.Store(false)
	if r := recover(); r != nil {
		// 这里的 *err 赋值是安全的
		*err = fmt.Errorf("%+v", r)
		slog.Info("请求主题失败", "err", *err)
	}
}

func (api *TopicGroupApi) makeResult(
	total int,
	page int,
) *response.GroupTopic {
	return &response.GroupTopic{
		Items:      lo.Subset(api.cacheData, (page-1)*perPage, perPage),
		Pagination: response.NewPerTenPageInfo(total, page),
	}
}

func (api *TopicGroupApi) groupRequestData(groupNode g.GroupNode) (
	[]response.TopicResult,
	*response.PerTenPageInfo,
	error,
) {
	// 并发获取子节点的数据
	var (
		ctx      = context.Background()
		chResult = make(chan *response.GroupTopic)
		chError  = make(chan error, 1)
		eg       errgroup.Group
	)
	slog.Info("正在请求主题", slog.String("node", groupNode.Name), slog.Int("len", len(groupNode.Nodes)))
	for _, nodeKey := range groupNode.Nodes {
		// 错误统一返回
		eg.Go(
			func() error {
				resp, err := api.requestTopics(ctx, nodeKey)
				if err == nil {
					chResult <- resp
				}

				// 如果错误是么有更多数据, 那么不返回
				slog.Info("请求", slog.String("node", nodeKey), slog.Any("err", err))
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
		pp      response.PerTenPageInfo
	)
	for result := range chResult {
		results = append(results, result.Items...)
		pp.TotalCount += result.Pagination.TotalCount
	}

	// 等待所有任务完成
	if err := <-chError; err != nil {
		return nil, nil, err
	}

	// 按照更新时间排序不同分组的数据
	slices.SortFunc(
		results, func(a, b response.TopicResult) int {
			return cmp.Compare(b.LastTouched, a.LastTouched)
		},
	)

	return results, &pp, nil
}
