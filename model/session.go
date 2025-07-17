package model

import (
	"sync/atomic"
)

type SessionData struct {
	HideFooter atomic.Bool
	IsApiV2    atomic.Bool
}
