package api

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dromara/carbon/v2"
	"github.com/seth-shi/go-v2ex/internal/types"
)

func (client *v2exClient) GetToken() tea.Msg {
	var res types.V2TokenResponse
	rr, err := client.client.R().
		SetContext(context.Background()).
		SetResult(&res).
		SetError(&res).
		Get("/api/v2/token")

	if err != nil {
		return err
	}

	if !res.Success {
		return fmt.Errorf("[%s]%s", rr.Status(), res.Message)
	}

	// 准备过期的话, 发送提醒
	expireAt := carbon.CreateFromTimestamp(res.Result.Created + res.Result.Expiration)
	if !carbon.Now().AddDays(14).Gte(expireAt) {
		return nil
	}

	return fmt.Errorf("token 将在%s过期,请注意更换", expireAt.String())
}
