package messages

import (
	"github.com/seth-shi/go-v2ex/internal/config"
)

type UiMessageInit struct {
	Config *config.FileConfig
	Error  error
}
