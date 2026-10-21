package api

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/response"
)

func (cli *v2exClient) GetReply(ctx context.Context, id int64, page int, ids ...uint64) tea.Cmd {
	reqID := requestID(ids)
	return func() tea.Msg {
		if !hasAuthToken() {
			res, err := cli.getPublicReplies(ctx, id, page)
			if err != nil {
				return messages.RequestError{Kind: messages.RequestReplies, TopicID: id, RequestID: reqID, Page: page, Err: errorWrapper("回复", err)}
			}
			return messages.GetReplyResponse{TopicID: id, RequestID: reqID, Page: page, Data: res}
		}

		var res response.V2ReplyResponse
		_, err := cli.client.R().
			SetContext(ctx).
			SetResult(&res).
			Get(fmt.Sprintf("/api/v2/topics/%d/replies?p=%d", id, page))

		if err != nil {
			fallback, fallbackErr := cli.getPublicReplies(ctx, id, page)
			if fallbackErr == nil {
				return messages.GetReplyResponse{TopicID: id, RequestID: reqID, Page: page, Data: fallback}
			}
			return messages.RequestError{Kind: messages.RequestReplies, TopicID: id, RequestID: reqID, Page: page, Err: errorWrapper("回复", errors.Join(err, fallbackErr))}
		}

		res.Pagination.CurrPage = page
		return messages.GetReplyResponse{
			TopicID:   id,
			RequestID: reqID,
			Page:      page,
			Data:      res,
		}
	}
}
