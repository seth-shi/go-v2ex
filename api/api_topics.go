package api

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/response"
)

func (cli *v2exClient) GetTopics(
	ctx context.Context,
	page int,
) tea.Cmd {

	return func() tea.Msg {

		var (
			conf       = g.Config.Get()
			nodeIndex  = conf.ActiveTab
			chooseV2   = conf.ChooseAPIV2
			node       = g.GetGroupNode(nodeIndex)
			res        []response.TopicResult
			total      int
			err        error
			cachePages = -1
		)

		// 如果是 myNodes, 那么就去用 V2 的接口
		v2 := chooseV2
		// 最新最热, 只能用 v1
		if node.Key == g.HotNode || node.Key == g.LatestNode {
			v2 = false
		}

		g.Session.IsApiV2.Store(v2)
		if v2 {
			res, cachePages, total, err = cli.v2TopicApi.GetTopicsByGroupNode(ctx, node, page)
		} else {
			res, total, err = cli.v1TopicApi.GetTopicsByGroupNode(ctx, node, page)
		}

		if err != nil {
			return errorWrapper("主题", err)
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
