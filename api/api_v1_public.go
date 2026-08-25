package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/pkg"
	"github.com/seth-shi/go-v2ex/v2/response"
)

var errPublicTopicNotFound = errors.New("公开接口未返回主题")

func (cli *v2exClient) getPublicDetail(ctx context.Context, id int64) (response.V2DetailResult, error) {
	var topics []response.TopicResult
	_, err := cli.client.R().
		SetContext(ctx).
		SetResult(&topics).
		Get(fmt.Sprintf("/api/topics/show.json?id=%d", id))
	if err != nil {
		return response.V2DetailResult{}, err
	}
	if len(topics) == 0 {
		return response.V2DetailResult{}, errPublicTopicNotFound
	}

	topic := topics[0]
	return response.V2DetailResult{
		Id:           topic.Id,
		Title:        topic.Title,
		Content:      topic.Content,
		Url:          topic.Url,
		Replies:      topic.Replies,
		LastReplyBy:  topic.LastReplyBy,
		Created:      topic.Created,
		LastModified: topic.LastModified,
		LastTouched:  topic.LastTouched,
		Member:       topic.Member,
		Node:         topic.Node,
	}, nil
}

func (cli *v2exClient) getPublicReplies(ctx context.Context, id int64, page int) (response.V2ReplyResponse, error) {
	var replies []response.V2ReplyResult
	_, err := cli.client.R().
		SetContext(ctx).
		SetResult(&replies).
		Get(fmt.Sprintf("/api/replies/show.json?topic_id=%d", id))
	if err != nil {
		return response.V2ReplyResponse{}, err
	}

	totalPages := pkg.TotalPages(len(replies), consts.PerPage)
	return response.V2ReplyResponse{
		Result: lo.Subset(replies, (page-1)*consts.PerPage, consts.PerPage),
		Pagination: response.V2PageResponse{
			TotalCount: len(replies),
			TotalPages: totalPages,
			CurrPage:   page,
		},
	}, nil
}
