package messages

import (
	"github.com/seth-shi/go-v2ex/internal/resources"
)

type GetTopics struct {
	Topics []*resources.TopicResource
	Error  error
}
