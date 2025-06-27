package messages

import (
	"github.com/seth-shi/go-v2ex/internal/model/response"
)

type GetTopicsRequest struct {
	Page int
}

type GetTopicResponse struct {
	Data *response.GroupTopic
}
