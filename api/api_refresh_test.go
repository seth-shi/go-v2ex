package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/api/internal/api_topics"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/pkg"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/stretchr/testify/require"
)

func newRequestTestClient(t *testing.T, handler http.HandlerFunc) *v2exClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := pkg.NewHTTPClient(&model.FileConfig{Timeout: 2}).
		SetBaseURL(server.URL).
		AddResponseMiddleware(apiErrorHandler)
	return &v2exClient{client: client, v1TopicApi: api_topics.NewV1(client), v2TopicApi: api_topics.NewV2(client)}
}

func setRequestTestConfig(t *testing.T, conf *model.FileConfig) {
	t.Helper()
	original := g.Config.Get()
	g.Config.Set(conf)
	t.Cleanup(func() { g.Config.Set(original) })
}

func TestTopicsCaptureNodeAndRefreshBypassesPublicCache(t *testing.T) {
	setRequestTestConfig(t, &model.FileConfig{ActiveTab: 7})
	var requests atomic.Int64
	cli := newRequestTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "qna", r.URL.Query().Get("node_name"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `[{"id":%d,"title":"topic","last_touched":100}]`, requests.Add(1))
	})

	firstCmd := cli.GetTopics(context.Background(), 1, 41)
	// Commands retain their original target even when navigation changes
	// before Bubble Tea executes them.
	g.Config.Set(&model.FileConfig{ActiveTab: 8})
	first, ok := firstCmd().(messages.GetTopicResponse)
	require.True(t, ok)
	require.Equal(t, "qna", first.NodeKey)
	require.Equal(t, uint64(41), first.RequestID)
	require.Equal(t, int64(1), first.Data[0].Id)

	node := g.GetGroupNode(7)
	cached := cli.GetTopicsForNode(context.Background(), node, 1, 42, false)().(messages.GetTopicResponse)
	require.Equal(t, int64(1), cached.Data[0].Id)
	require.Equal(t, int64(1), requests.Load())

	g.Config.Set(&model.FileConfig{ActiveTab: 7})
	fresh := cli.GetTopicsFresh(context.Background(), 1, 43)().(messages.GetTopicResponse)
	require.Equal(t, int64(2), fresh.Data[0].Id)
	require.Equal(t, uint64(43), fresh.RequestID)
	require.Equal(t, int64(2), requests.Load())
}

func TestRefreshCannotBeOverwrittenByOlderInflightTopics(t *testing.T) {
	setRequestTestConfig(t, &model.FileConfig{ActiveTab: 7})
	started, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	var requests atomic.Int64
	cli := newRequestTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		id := requests.Add(1)
		if id == 1 {
			close(started)
			<-release
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `[{"id":%d,"title":"topic","last_touched":100}]`, id)
	})

	oldResults, freshResults := make(chan tea.Msg, 1), make(chan tea.Msg, 1)
	oldCmd := cli.GetTopics(context.Background(), 1, 1)
	freshCmd := cli.GetTopicsFresh(context.Background(), 1, 2)
	go func() { oldResults <- oldCmd() }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("the initial request did not start")
	}
	go func() { freshResults <- freshCmd() }()
	releaseOnce.Do(func() { close(release) })

	for _, result := range []struct {
		ch <-chan tea.Msg
		id uint64
	}{
		{oldResults, 1}, {freshResults, 2},
	} {
		select {
		case msg := <-result.ch:
			topics, ok := msg.(messages.GetTopicResponse)
			require.True(t, ok, "unexpected result %T", msg)
			require.Equal(t, result.id, topics.RequestID)
			require.Equal(t, int64(result.id), topics.Data[0].Id)
		case <-time.After(3 * time.Second):
			t.Fatal("topic request did not complete")
		}
	}
	cached := cli.GetTopics(context.Background(), 1, 3)().(messages.GetTopicResponse)
	require.Equal(t, int64(2), cached.Data[0].Id)
	require.Equal(t, int64(2), requests.Load())
}

func TestFreshV2TopicsReloadCurrentDeepPageAndResetCursors(t *testing.T) {
	setRequestTestConfig(t, &model.FileConfig{Token: "test-only-token", ActiveTab: 7})
	var generation, firstPages, secondPages atomic.Int64
	cli := newRequestTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/topics/show.json" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		if r.URL.Path != "/api/v2/nodes/qna/topics" {
			http.NotFound(w, r)
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("p"))
		if page == 1 {
			generation.Add(1)
			firstPages.Add(1)
		} else if page == 2 {
			secondPages.Add(1)
		} else {
			t.Errorf("unexpected V2 cursor page: %d", page)
		}
		items := make([]response.V2TopicResult, 20)
		for i := range items {
			id := generation.Load()*100 + int64((page-1)*20+i+1)
			items[i] = response.V2TopicResult{Id: id, Title: "topic", LastTouched: 1000 - int64(i)}
		}
		_ = json.NewEncoder(w).Encode(response.V2TopicResponse{
			Result: items, Pagination: response.V2PageResponse{TotalCount: 40, TotalPages: 2},
		})
	})

	for i := 1; i <= 2; i++ {
		msg := cli.GetTopicsFresh(context.Background(), 3, uint64(i))()
		topics, ok := msg.(messages.GetTopicResponse)
		require.True(t, ok, "unexpected result %T", msg)
		require.Len(t, topics.Data, 10)
		require.Equal(t, int64(i*100+21), topics.Data[0].Id)
		require.Equal(t, 3, topics.PageInfo.CurrPage)
		require.Equal(t, 40, topics.PageInfo.TotalCount)
	}
	require.Equal(t, int64(2), firstPages.Load())
	require.Equal(t, int64(2), secondPages.Load())
}

func TestRequestFailuresKeepTheirPageAndRequestScope(t *testing.T) {
	setRequestTestConfig(t, &model.FileConfig{ActiveTab: 7})
	cli := newRequestTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "temporarily unavailable", http.StatusServiceUnavailable)
	})
	cases := []struct {
		cmd   tea.Cmd
		kind  messages.RequestKind
		node  string
		topic int64
		id    uint64
		page  int
	}{
		{cli.GetTopics(context.Background(), 3, 9), messages.RequestTopics, "qna", 0, 9, 3},
		{cli.GetDetail(context.Background(), 42, 10), messages.RequestDetail, "", 42, 10, 0},
		{cli.GetReply(context.Background(), 42, 2, 11), messages.RequestReplies, "", 42, 11, 2},
	}
	for _, tc := range cases {
		msg := tc.cmd()
		err, ok := msg.(messages.RequestError)
		require.True(t, ok, "unexpected result %T", msg)
		require.Equal(t, tc.kind, err.Kind)
		require.Equal(t, tc.node, err.NodeKey)
		require.Equal(t, tc.topic, err.TopicID)
		require.Equal(t, tc.id, err.RequestID)
		require.Equal(t, tc.page, err.Page)
		require.NotNil(t, err.Unwrap())
		require.Contains(t, err.Error(), "503")
	}
}

func TestV2GroupFailureDoesNotAdvanceSuccessfulPeer(t *testing.T) {
	var mu sync.Mutex
	requests := make(map[string]int)
	cli := newRequestTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		nodeName := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/nodes/"), "/topics")
		page, _ := strconv.Atoi(r.URL.Query().Get("p"))
		key := fmt.Sprintf("%s/%d", nodeName, page)
		mu.Lock()
		requests[key]++
		attempt := requests[key]
		mu.Unlock()
		if nodeName == "second" && page == 2 && attempt == 1 {
			http.Error(w, "temporary failure", http.StatusServiceUnavailable)
			return
		}
		base := 100
		if nodeName == "second" {
			base = 200
		}
		items := make([]response.V2TopicResult, 5)
		for i := range items {
			items[i] = response.V2TopicResult{Id: int64(base + page*10 + i), Title: key, LastTouched: int64(100 - page*10 - i)}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response.V2TopicResponse{
			Result: items, Pagination: response.V2PageResponse{TotalCount: 10, TotalPages: 2},
		})
	})
	node := g.GroupNode{Key: "transactional-test", Nodes: []string{"first", "second"}}
	firstPage, _, total, err := cli.v2TopicApi.GetTopicsByGroupNode(context.Background(), node, 1)
	require.NoError(t, err)
	require.Len(t, firstPage, 10)
	require.Equal(t, 20, total)

	_, _, _, err = cli.v2TopicApi.GetTopicsByGroupNode(context.Background(), node, 2)
	require.Error(t, err)
	// A failed later batch must leave the already published cache intact.
	cached, _, _, err := cli.v2TopicApi.GetTopicsByGroupNode(context.Background(), node, 1)
	require.NoError(t, err)
	require.Equal(t, firstPage, cached)

	retried, _, total, err := cli.v2TopicApi.GetTopicsByGroupNode(context.Background(), node, 2)
	require.NoError(t, err)
	require.Equal(t, 20, total)
	require.Len(t, retried, 10)
	ids := make([]int64, len(retried))
	for i, topic := range retried {
		ids[i] = topic.Id
	}
	require.ElementsMatch(t, []int64{120, 121, 122, 123, 124, 220, 221, 222, 223, 224}, ids)
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 1, requests["first/1"])
	require.Equal(t, 1, requests["second/1"])
	require.Equal(t, 2, requests["first/2"], "the successful peer must retry the discarded page")
	require.Equal(t, 2, requests["second/2"])
}
