package http

import (
	"context"
	"errors"
	"log"

	"github.com/seth-shi/go-v2ex/internal/resources"
)

type MemberResponse struct {
	apiError
	Result *resources.MemberResult `json:"result"`
}

func (client *v2exClient) GetMember(ctx context.Context) (*resources.MemberResult, error) {
	var res MemberResponse
	rrr, err := client.client.R().
		SetContext(ctx).
		SetResult(&res).
		SetError(&res).
		Get("/api/v2/member")
	if err != nil {
		return nil, err
	}

	log.Println(rrr.Request.Header)
	if !res.Success {
		return nil, errors.New(res.Message)
	}

	return res.Result, nil
}
