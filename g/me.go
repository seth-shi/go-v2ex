package g

import (
	"github.com/seth-shi/go-v2ex/model"
	"github.com/seth-shi/go-v2ex/response"
)

var (
	Me = model.
		NewSafe(response.MeResult{}, nil)
)
