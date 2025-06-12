package api

import (
	"context"
	"log"

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

		res, err := cli.requestTopics(ctx, nodeIndex, page)
		if err != nil {
			return errorWrapper("主题", err)
		}

		// 此次成功的话预加载下一页和左右 tab 的
		go func() {
			_, err := cli.requestTopics(ctx, nodeIndex, page+1)
			if err != nil {
				log.Printf("请求主题失败: %v", err)
			}
		}()

		return messages.GetTopicResponse{Data: res}
	}
}

func (cli *v2exClient) requestTopics(ctx context.Context, nodeIndex int, page int) (*response.Topic, error) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("请求主题失败: %v", r)
		}
	}()

	if page <= 0 {
		page = 1
	}
	var (
		nodeName = lo.NthOr(config.G.GetNodes(), nodeIndex, latestNode)
		res      *response.Topic
		err      error
	)
	switch nodeName {
	case latestNode, hotNode:
		res, err = cli.getV1Topics(ctx, nodeName, page)
	default:
		res, err = cli.getV2Topics(ctx, nodeName, page)
	}
	return res, err
}
