package g

import (
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/response"
)

var (
	Me = model.
		NewSafe(response.MeResult{}, nil)
)
