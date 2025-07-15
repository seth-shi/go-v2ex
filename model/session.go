package model

import (
	"sync/atomic"
)

type SessionData struct {
	HideFooter  atomic.Bool
	ChooseApiV2 atomic.Bool
	IsApiV2     atomic.Bool
}
