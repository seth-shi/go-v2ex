package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/seth-shi/go-v2ex/v2/api/internal/api_topics"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/pkg"
	"github.com/stretchr/testify/require"
)

func TestAnonymousDetailAndRepliesUsePublicAPI(t *testing.T) {
	originalConfig := g.Config.Get()
	g.Config.Set(&model.FileConfig{Timeout: 1, ActiveTab: 7})
	t.Cleanup(func() { g.Config.Set(originalConfig) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v2/") {
			t.Errorf("anonymous request must not use V2 API: %s", r.URL.String())
		}
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("public API request must not include authorization, got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/topics/show.json":
			_, _ = w.Write([]byte(`[{"id":42,"title":"公开主题","content":"正文","url":"https://www.v2ex.com/t/42","replies":12,"last_reply_by":"bob","created":100,"last_modified":200,"last_touched":210,"member":{"id":1,"username":"alice"},"node":{"id":2,"name":"qna","title":"问与答"}}]`))
		case "/api/replies/show.json":
			_, _ = w.Write([]byte(publicRepliesJSON(12)))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client := pkg.NewHTTPClient(&model.FileConfig{Timeout: 1}).
		SetBaseURL(server.URL).
		AddRequestMiddleware(beforeRequest).
		AddResponseMiddleware(apiErrorHandler)
	cli := &v2exClient{
		client:     client,
		v1TopicApi: api_topics.NewV1(client),
		v2TopicApi: api_topics.NewV2(client),
	}

	topicsMsg := cli.GetTopics(context.Background(), 1)()
	topics, ok := topicsMsg.(messages.GetTopicResponse)
	require.True(t, ok, "unexpected topics message: %T", topicsMsg)
	require.Len(t, topics.Data, 1)
	require.Equal(t, int64(42), topics.Data[0].Id)

	detailMsg := cli.GetDetail(context.Background(), 42)()
	detail, ok := detailMsg.(messages.GetDetailResponse)
	require.True(t, ok, "unexpected detail message: %T", detailMsg)
	require.Equal(t, int64(42), detail.Data.Id)
	require.Equal(t, "alice", detail.Data.Member.Username)
	require.Equal(t, "正文", detail.Data.Content)

	firstMsg := cli.GetReply(context.Background(), 42, 1)()
	first, ok := firstMsg.(messages.GetReplyResponse)
	require.True(t, ok, "unexpected reply message: %T", firstMsg)
	require.Len(t, first.Data.Result, 10)
	require.Equal(t, 12, first.Data.Pagination.TotalCount)
	require.Equal(t, 2, first.Data.Pagination.TotalPages)
	require.Equal(t, 1, first.Data.Pagination.CurrPage)

	secondMsg := cli.GetReply(context.Background(), 42, 2)()
	second, ok := secondMsg.(messages.GetReplyResponse)
	require.True(t, ok, "unexpected reply message: %T", secondMsg)
	require.Len(t, second.Data.Result, 2)
	require.Equal(t, 2, second.Data.Pagination.CurrPage)
}

func publicRepliesJSON(count int) string {
	items := make([]string, count)
	for i := range items {
		items[i] = `{"id":1,"content":"回复","created":101,"member":{"id":2,"username":"bob"}}`
	}
	return "[" + strings.Join(items, ",") + "]"
}
