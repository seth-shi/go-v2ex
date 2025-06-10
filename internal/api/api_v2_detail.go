package api

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

func (client *v2exClient) GetDetail(ctx context.Context, id int64) tea.Cmd {
	return func() tea.Msg {

		var res response.V2Detail
		rr, err := client.client.R().
			SetContext(ctx).
			SetResult(&res).
			SetError(&res).
			Get(fmt.Sprintf("/api/v2/topics/%d", id))

		if err != nil {
			return errorWrapper("详情", err)
		}

		if !res.IsSuccess() {
			return errorWrapper("详情", fmt.Errorf("[%s]%s", rr.Status(), res.Message))
		}

		return messages.GetDetailResponse{Data: res.Result}
	}
}
