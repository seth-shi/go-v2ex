package events

import (
	ctx "context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/http"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

func GetMe() tea.Msg {
	member, err := http.V2exClient.GetMember(ctx.Background())
	return messages.GetMe{
		Member: member,
		Error:  err,
	}
}
