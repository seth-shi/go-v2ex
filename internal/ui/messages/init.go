package messages

import (
	"github.com/seth-shi/go-v2ex/internal/types"
)

type LoadConfigRequest struct {
}

type LoadConfigResult struct {
	Config types.FileConfig
	Error  error
}
