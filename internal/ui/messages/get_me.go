package messages

import (
	"github.com/seth-shi/go-v2ex/internal/types"
)

type GetMeRequest struct {
}

type GetMeResult struct {
	Member *types.V2MemberResult
	Error  error
}
