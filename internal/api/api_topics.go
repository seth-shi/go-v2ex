package api

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
)

func (cli *v2exClient) GetTopics(
	ctx context.Context,
	nodeIndex int,
	page int,
) tea.Cmd {

	return func() tea.Msg {

		var (
			node = config.GetGroupNode(nodeIndex)
		)

		res, err := cli.topicApi.GetTopicsByGroupNode(ctx, node, page)
		if err != nil {
			return errorWrapper("主题", err)
		}

		return messages.GetTopicResponse{Data: res}
	}
}
