package api

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/g"
	"github.com/seth-shi/go-v2ex/response"
)

func (cli *v2exClient) Me(ctx context.Context) tea.Cmd {
	return func() tea.Msg {

		var res response.MeResponse
		_, err := cli.client.R().
			SetContext(ctx).
			SetResult(&res).
			Get("/api/v2/member")
		if err != nil {
			return err
		}

		g.Me.Set(res.Result)

		return nil
	}
}
