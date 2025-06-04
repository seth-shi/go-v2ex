package api

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

func (client *v2exClient) GetReply(id int64, page int) tea.Cmd {
	return func() tea.Msg {
		var res types.V2ReplyResponse
		rr, err := client.client.R().
			SetContext(context.Background()).
			SetResult(&res).
			SetError(&res).
			Get(fmt.Sprintf("/api/v2/topics/%d/replies?p=%d", id, page))

		if err != nil {
			return messages.GetRepliesResult{Error: err}
		}

		if !res.Success {
			return messages.GetRepliesResult{Error: fmt.Errorf("[%s]%s", rr.Status(), res.Message)}
		}

		res.Pagination.CurrPage = page
		return messages.GetRepliesResult{Replies: res.Result, Pagination: res.Pagination}
	}
}
