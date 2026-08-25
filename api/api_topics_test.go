package api

import (
	"testing"

	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/stretchr/testify/require"
)

func TestMergeTopicMetadataByTopicID(t *testing.T) {
	canonical := []response.TopicResult{
		{Id: 1, Title: "one", Replies: 5, LastReplyBy: "reply_user"},
		{Id: 2, Title: "two", LastReplyBy: "second_reply"},
	}
	metadata := []response.TopicResult{
		{
			Id:      1,
			Visits:  12_345,
			Member:  response.MemberResult{Id: 7, Username: "author"},
			Node:    response.NodeInfoResult{Title: "程序员"},
			Created: 100,
		},
	}

	merged := mergeTopicMetadata(canonical, metadata)
	require.Equal(t, "author", merged[0].Member.Username)
	require.Equal(t, "reply_user", merged[0].LastReplyBy)
	require.Equal(t, 12_345, merged[0].Visits)
	require.Equal(t, "程序员", merged[0].Node.Title)
	require.Empty(t, merged[1].Member.Username)
	require.Equal(t, "second_reply", merged[1].LastReplyBy)
}
