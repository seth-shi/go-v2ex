package events

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/resources"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

// 声明周期请求生成一次即可

func GetTopics(page int) tea.Cmd {
	return func() tea.Msg {
		var topics []*resources.TopicResource
		for i := 0; i < 10; i++ {
			topic := resources.GetDefaultTopic()
			topic.Title = fmt.Sprintf("第%d页-%s", page, topic.Title)
			topics = append(topics, topic)
			time.Sleep(time.Millisecond * 10)
		}
		return messages.GetTopics{
			Topics: topics,
			Error:  nil,
		}
	}
}
