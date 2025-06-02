package api

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

func (client *v2exClient) GetDetail(id int64) tea.Cmd {
	return func() tea.Msg {
		var res types.V2DetailResponse
		rr, err := client.client.R().
			SetContext(context.Background()).
			SetResult(&res).
			SetError(&res).
			Get(fmt.Sprintf("/api/v2/topics/%d", id))

		if err != nil {
			return err
		}

		if !res.Success {
			return fmt.Errorf("[%s]%s", rr.Status(), res.Message)
		}

		return messages.GetDetailResult{Detail: res.Result}
	}
}
