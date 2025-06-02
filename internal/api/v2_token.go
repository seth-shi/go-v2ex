package api

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dromara/carbon/v2"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

func (client *v2exClient) GetToken() tea.Msg {
	var res types.V2TokenResponse
	_, err := client.client.R().
		SetContext(context.Background()).
		SetResult(&res).
		SetError(&res).
		Get("/api/v2/token")

	if err != nil {
		return err
	}

	if !res.Success {
		return errors.New(res.Message)
	}

	return messages.ShowAutoClearTipsRequest{Text: fmt.Sprintf("token 有效期: %s", carbon.CreateFromTimestamp(res.Result.Created+res.Result.Expiration).String())}
}
