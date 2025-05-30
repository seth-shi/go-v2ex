package events

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

func InitFileConfig() tea.Msg {
	cfg, err := config.NewConfig()
	time.Sleep(time.Second * 1)
	return messages.UiMessageInit{
		Config: cfg,
		Error:  err,
	}
}
