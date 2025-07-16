package g

import (
	"github.com/seth-shi/go-v2ex/v2/model"
)

var (
	Config = model.
		NewSafe(model.NewDefaultFileConfig(), model.SaveToFile)
)
