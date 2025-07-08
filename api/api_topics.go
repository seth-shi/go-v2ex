package api

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/g"
	"github.com/seth-shi/go-v2ex/messages"
)

func (cli *v2exClient) GetTopics(
	ctx context.Context,
	page int,
) tea.Cmd {

	return func() tea.Msg {

		var (
			nodeIndex = g.Config.Get().ActiveTab
			node      = g.GetGroupNode(nodeIndex)
		)

		res, err := cli.topicApi.GetTopicsByGroupNode(ctx, node, page)
		if err != nil {
			return errorWrapper("主题", err)
		}

		return messages.GetTopicResponse{Data: res}
	}
}
