package messages

import (
	"github.com/seth-shi/go-v2ex/v2/model"
)

type LoadConfigResult struct {
	Result *model.FileConfig
	Err    error
}
