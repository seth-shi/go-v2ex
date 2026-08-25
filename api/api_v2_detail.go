package api

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/response"
)

func (cli *v2exClient) GetDetail(ctx context.Context, id int64) tea.Cmd {
	return func() tea.Msg {
		if !hasAuthToken() {
			res, err := cli.getPublicDetail(ctx, id)
			if err != nil {
				return errorWrapper("详情", err)
			}
			return messages.GetDetailResponse{Data: res}
		}

		var res response.V2Detail
		_, err := cli.client.R().
			SetContext(ctx).
			SetResult(&res).
			Get(fmt.Sprintf("/api/v2/topics/%d", id))
		if err != nil {
			fallback, fallbackErr := cli.getPublicDetail(ctx, id)
			if fallbackErr == nil {
				return messages.GetDetailResponse{Data: fallback}
			}
			return errorWrapper("详情", errors.Join(err, fallbackErr))
		}

		return messages.GetDetailResponse{Data: res.Result}
	}
}
