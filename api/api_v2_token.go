package api

import (
	"context"
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dromara/carbon/v2"
	"github.com/seth-shi/go-v2ex/response"
)

var (
	tokenOnce sync.Once
	tokenMsg  error
)

func (cli *v2exClient) GetToken(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		tokenOnce.Do(
			func() {
				var res response.V2Token
				_, err := cli.client.R().
					SetContext(ctx).
					SetResult(&res).
					Get("/api/v2/token")

				if err != nil {
					tokenMsg = err
					return
				}

				// 准备过期的话, 发送提醒
				expireAt := carbon.CreateFromTimestamp(res.Result.Created + res.Result.Expiration)
				if !carbon.Now().AddDays(14).Gte(expireAt) {
					return
				}

				tokenMsg = errorWrapper("令牌", fmt.Errorf("将在%s过期,请注意更换", expireAt.String()))
			},
		)

		return tokenMsg
	}
}
