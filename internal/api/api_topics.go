package api

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

func (cli *v2exClient) GetTopics(
	ctx context.Context,
	nodeIndex int,
	page int,
) tea.Cmd {

	return func() tea.Msg {

		if page <= 0 {
			page = 1
		}

		var (
			nodeName = lo.NthOr(config.G.GetNodes(), nodeIndex, latestNode)
			res      *response.Topic
			err      error
		)

		// 请求的时候, 用数据的分页数据
		switch nodeName {
		case latestNode, hotNode:
			res, err = cli.getV1Topics(ctx, nodeName, page)
		default:
			res, err = cli.getV2Topics(ctx, nodeName, page)
		}

		if err != nil {
			return errorWrapper("主题", err)
		}

		return messages.GetTopicResponse{Data: res}
	}
}
