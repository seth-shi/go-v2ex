package api

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

func (cli *v2exClient) GetReply(ctx context.Context, id int64, page int) tea.Cmd {
	return func() tea.Msg {

		var res response.V2Reply
		_, err := cli.client.R().
			SetContext(ctx).
			SetResult(&res).
			Get(fmt.Sprintf("/api/v2/topics/%d/replies?p=%d", id, page))

		if err != nil {
			return errorWrapper("回复", err)
		}

		res.Pagination.CurrPage = page
		return messages.GetReplyResponse{Data: res}
	}
}
