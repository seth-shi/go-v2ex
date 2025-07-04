package messages

import (
	"github.com/seth-shi/go-v2ex/internal/config"
)

type LoadConfigResult struct {
	Result *config.FileConfig
}
