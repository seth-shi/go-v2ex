package api

import (
	"context"
	"errors"
	"fmt"
	"log"

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
	otherUri    = "/api/topics/show.json?node_name=%s&p=%d"
)

func (client *v2exClient) GetTopics(nodeIndex int, page int) tea.Cmd {

	return func() tea.Msg {

		if page <= 0 {
			page = 1
		}

		var (
			nodeName = lo.NthOr(config.G.GetNodes(), nodeIndex, latestNode)
			uri      string
		)
		switch nodeName {
		case latestNode:
			uri = latestUri
		case hotNode:
			uri = hotUri
		default:
			uri = fmt.Sprintf(otherUri, nodeName, page)
		}

		var apiErr types.V1ApiError
		var res []types.V1TopicResult
		_, err := client.client.R().
			SetContext(context.Background()).
			SetResult(&res).
			SetError(&apiErr).
			Get(uri)

		if !apiErr.Success() {
			err = errors.New(apiErr.Message)
		}

		// 最新无分页, 手动返回翻页
		if uri == latestUri {
			offset := (page - 1) * latestLimit
			log.Println(offset, " ", len(res))
			res = lo.Slice(res, offset, offset+latestLimit)
		}

		return messages.GetTopicsResult{
			Topics: res,
			Page:   page,
			Error:  err,
		}
	}

}
