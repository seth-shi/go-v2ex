package api

import (
	"fmt"
	"time"

	"github.com/seth-shi/go-v2ex/internal/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

func GetTopics(page int) tea.Cmd {
	return func() tea.Msg {

		time.Sleep(time.Second * 1)
		var topics []*types.TopicResource
		for i := 0; i < 10; i++ {
			topic := types.GetDefaultTopic()
			topic.Title = fmt.Sprintf("第%d页-%s", page, topic.Title)
			topics = append(topics, topic)
			time.Sleep(time.Millisecond * 10)
		}
		return messages.GetTopicsResult{
			Topics: topics,
			Error:  nil,
		}
	}
}
