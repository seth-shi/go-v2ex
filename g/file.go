package g

import (
	"github.com/seth-shi/go-v2ex/model"
	"github.com/seth-shi/go-v2ex/pkg"
)

var (
	Config = pkg.
		NewSafe(model.NewDefaultFileConfig(), model.SaveToFile)
)
