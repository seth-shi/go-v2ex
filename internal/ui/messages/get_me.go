package messages

import (
	"github.com/seth-shi/go-v2ex/internal/resources"
)

type GetMe struct {
	Member *resources.MemberResult
	Error  error
}
