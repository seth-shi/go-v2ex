package api

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

func (client *v2exClient) GetMember() tea.Msg {
	var res types.V2MemberResponse
	_, err := client.client.R().
		SetContext(context.Background()).
		SetResult(&res).
		SetError(&res).
		Get("/api/v2/member")

	result := messages.GetMeResult{}
	if err != nil {
		result.Error = err
		return result
	}

	if !res.Success {
		result.Error = errors.New(res.Message)
		return result
	}

	result.Member = res.Result
	return result
}
