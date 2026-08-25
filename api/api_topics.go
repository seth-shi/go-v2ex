package api

import (
	"context"
	"errors"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/pkg"
	"github.com/seth-shi/go-v2ex/v2/response"
)

func (cli *v2exClient) GetTopics(
	ctx context.Context,
	page int,
) tea.Cmd {

	return func() tea.Msg {
		var (
			node       = g.GetGroupNode(g.Config.Get().ActiveTab)
			res        []response.TopicResult
			total      int
			cachePages = -1
		)

		// The hot/latest endpoints only exist in V1. All other groups use V2 as
		// the canonical paginated source and V1 as rich metadata enrichment.
		if !hasAuthToken() || node.Key == g.HotNode || node.Key == g.LatestNode {
			var err error
			res, total, err = cli.v1TopicApi.GetTopicsByGroupNode(ctx, node, page)
			if err != nil {
				return errorWrapper("主题", err)
			}
		} else {
			var err error
			res, cachePages, total, err = cli.getHybridTopics(ctx, node, page)
			if err != nil {
				return errorWrapper("主题", err)
			}
		}

		return messages.GetTopicResponse{
			Data: res,
			PageInfo: &response.PerTenPageInfo{
				TotalCount: total,
				CurrPage:   page,
			},
			CachePages: cachePages,
		}
	}
}

func (cli *v2exClient) getHybridTopics(ctx context.Context, node g.GroupNode, page int) (
	res []response.TopicResult, cachePages, total int, err error,
) {
	var (
		v1Topics         []response.TopicResult
		v2Topics         []response.TopicResult
		v1Total, v2Total int
		v2CachePages     int
		v1Err, v2Err     error
		wg               sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		v1Topics, v1Total, v1Err = cli.getAllV1Topics(ctx, node)
	}()
	go func() {
		defer wg.Done()
		v2Topics, v2CachePages, v2Total, v2Err = cli.v2TopicApi.GetTopicsByGroupNode(ctx, node, page)
	}()
	wg.Wait()

	if v2Err == nil {
		return mergeTopicMetadata(v2Topics, v1Topics), v2CachePages, v2Total, nil
	}
	if v1Err == nil {
		return lo.Subset(v1Topics, (page-1)*consts.PerPage, consts.PerPage), pkg.TotalPages(len(v1Topics), consts.PerPage), v1Total, nil
	}
	return nil, -1, 0, errors.Join(v2Err, v1Err)
}

func (cli *v2exClient) getAllV1Topics(ctx context.Context, node g.GroupNode) ([]response.TopicResult, int, error) {
	first, total, err := cli.v1TopicApi.GetTopicsByGroupNode(ctx, node, 1)
	if err != nil {
		return nil, 0, err
	}
	all := append([]response.TopicResult(nil), first...)
	for page := 2; page <= pkg.TotalPages(total, consts.PerPage); page++ {
		items, _, pageErr := cli.v1TopicApi.GetTopicsByGroupNode(ctx, node, page)
		if pageErr != nil {
			return all, total, pageErr
		}
		all = append(all, items...)
	}
	return all, total, nil
}

func mergeTopicMetadata(topics, metadata []response.TopicResult) []response.TopicResult {
	byID := lo.SliceToMap(metadata, func(topic response.TopicResult) (int64, response.TopicResult) {
		return topic.Id, topic
	})
	return lo.Map(topics, func(topic response.TopicResult, _ int) response.TopicResult {
		rich, ok := byID[topic.Id]
		if !ok {
			return topic
		}
		topic.Member = rich.Member
		topic.Visits = rich.Visits
		if rich.Node.Title != "" {
			topic.Node = rich.Node
		}
		if rich.Created > 0 {
			topic.Created = rich.Created
		}
		return topic
	})
}
