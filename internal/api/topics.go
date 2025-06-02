package api

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

const (
	latestNode  = "latest"
	latestUri   = "/api/topics/latest.json"
	latestLimit = 10
	hotNode     = "hot"
	hotUri      = "/api/topics/hot.json"
	v2TopicsUri = "/api/v2/nodes/%s/topics?p=%d"
)

func (client *v2exClient) GetTopics(nodeIndex int, page int) tea.Cmd {

	return func() tea.Msg {

		if page <= 0 {
			page = 1
		}

		var (
			nodeName = lo.NthOr(config.G.GetNodes(), nodeIndex, latestNode)
			uri      string
			v1Error  types.V1ApiError
			res      []types.TopicComResult
			v1Res    []types.V1TopicResult
			v2Res    types.V2TopicResponse
			err      error
		)

		r := client.client.R().SetContext(context.Background())
		switch nodeName {
		case latestNode, hotNode:
			uri = lo.If(nodeName == latestNode, latestUri).Else(hotUri)
			_, err = r.SetResult(&v1Res).SetError(&v1Error).Get(uri)
			if !v1Error.Success() {
				err = errors.New(v1Error.Message)
			}

			// 这两个接口不支持分页, 手动切分
			offset := (page - 1) * latestLimit
			v1Res = lo.Slice(v1Res, offset, offset+latestLimit)
			// 转换成统一的输出
			res = lo.Map(v1Res, func(item types.V1TopicResult, index int) types.TopicComResult {
				return types.TopicComResult{
					Id:          item.Id,
					Node:        item.Node.Title,
					Title:       item.Title,
					Member:      item.Member.Username,
					LastTouched: item.LastTouched,
					Replies:     item.Replies,
				}
			})
		default:
			// 使用 V2 的接口
			uri = fmt.Sprintf(v2TopicsUri, nodeName, page)
			_, err = r.SetResult(&v2Res).SetError(&v2Res).Get(uri)
			if !v2Res.Success {
				err = errors.New(v2Res.Message)
			}
			res = lo.Map(v2Res.Result, func(item types.V2TopicResult, index int) types.TopicComResult {
				return types.TopicComResult{
					Id:          item.Id,
					Node:        nodeName,
					Title:       item.Title,
					Member:      item.LastReplyBy,
					LastTouched: item.LastTouched,
					Replies:     item.Replies,
				}
			})
		}

		return messages.GetTopicsResult{
			Topics: res,
			Page:   page,
			Error:  err,
		}
	}
}
