package pages

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/stretchr/testify/require"
)

// These tests deliver response messages directly. Returned API commands are
// inspected but never executed, so no account or real V2EX endpoint is used.
func requestPageEnvironment(t *testing.T) string {
	t.Helper()
	originalHistory, originalConfig := g.History, g.Config.Get()
	w, h := g.Window.GetSize()
	hideFooter := g.Session.HideFooter.Load()
	path := filepath.Join(t.TempDir(), "history.json")
	g.History = model.NewHistoryStore(path)
	require.NoError(t, g.History.Load())
	g.Config.Set(&model.FileConfig{ActiveTab: 7})
	g.Window.SetSize(tea.WindowSizeMsg{Width: 100, Height: 30})
	g.Session.HideFooter.Store(false)
	t.Cleanup(func() {
		g.History = originalHistory
		g.Config.Set(originalConfig)
		g.Window.SetSize(tea.WindowSizeMsg{Width: w, Height: h})
		g.Session.HideFooter.Store(hideFooter)
	})
	return path
}

func requestTestTopics() []response.TopicResult {
	return []response.TopicResult{
		{Id: 1, Title: "第一篇主题"},
		{Id: 2, Title: "当前选中的主题"},
		{Id: 3, Title: "第三篇主题"},
	}
}

func loadedRequestTopicPage() topicPage {
	page := newTopicPage()
	page.topics = requestTestTopics()
	page.index, page.page, page.pendingPage, page.totalPages = 1, 2, 2, 3
	page.loading = false
	return page
}

func loadedRequestDetailPage() detailPage {
	page := newDetailPage(42)
	page.loading = false
	page.contentDetail = response.V2DetailResult{
		Id: 42, Title: "已加载的正文", Content: "保留这段内容", Replies: 30,
		Member: response.MemberResult{Id: 1, Username: "author"},
		Node:   response.NodeInfoResult{Title: "程序员"},
	}
	page.contentReply = []response.V2ReplyResult{{Id: 1, Content: "第一条评论"}, {Id: 2, Content: "第二条评论"}}
	page.replyPageInfo = response.V2PageResponse{CurrPage: 1, TotalPages: 3, TotalCount: 30}
	return page
}

func TestInitialTopicFailureShowsRetryAndStopsLoading(t *testing.T) {
	requestPageEnvironment(t)
	page := newTopicPage()
	failure := errors.New("模拟请求超时")
	updated, cmd := page.Update(messages.RequestError{
		Kind: messages.RequestTopics, NodeKey: page.node.Key, RequestID: page.requestID, Page: 1, Err: failure,
	})
	failed := updated.(topicPage)
	require.False(t, failed.loading)
	require.ErrorIs(t, failed.loadErr, failure)
	require.Contains(t, failed.View(), "主题加载失败")
	require.Contains(t, failed.View(), "Ctrl+R")
	require.Nil(t, cmd)

	updated, retry := failed.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	retrying := updated.(topicPage)
	require.NotNil(t, retry)
	require.True(t, retrying.loading)
	require.True(t, retrying.fresh)
	require.Nil(t, retrying.loadErr)
	require.NotEqual(t, failed.requestID, retrying.requestID)
	require.Equal(t, 1, retrying.pendingPage)
}

func TestTopicRefreshFailureRetainsExistingListAndSelection(t *testing.T) {
	requestPageEnvironment(t)
	page := loadedRequestTopicPage()
	updated, refresh := page.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	refreshing := updated.(topicPage)
	require.NotNil(t, refresh)
	require.True(t, refreshing.loading)
	require.Equal(t, page.topics, refreshing.topics)
	require.Equal(t, page.index, refreshing.index)
	require.Equal(t, 2, refreshing.pendingPage)

	failure := errors.New("刷新暂时不可用")
	updated, _ = refreshing.Update(messages.RequestError{
		Kind: messages.RequestTopics, NodeKey: page.node.Key, RequestID: refreshing.requestID, Page: 2, Err: failure,
	})
	failed := updated.(topicPage)
	require.False(t, failed.loading)
	require.Equal(t, page.topics, failed.topics)
	require.Equal(t, page.index, failed.index)
	require.Equal(t, page.page, failed.page)
	require.Contains(t, failed.View(), "保留当前列表")
	require.Contains(t, failed.View(), "Ctrl+R")
}

func TestTopicRefreshKeepsSelectedTopicByIDOnSamePage(t *testing.T) {
	requestPageEnvironment(t)
	page := loadedRequestTopicPage()
	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	refreshing := updated.(topicPage)
	require.NotNil(t, cmd)
	updated, _ = refreshing.Update(messages.GetTopicResponse{
		NodeKey: page.node.Key, RequestID: refreshing.requestID,
		Data:     []response.TopicResult{{Id: 3, Title: "第三篇更新"}, {Id: 1, Title: "第一篇更新"}, {Id: 2, Title: "选中主题更新"}},
		PageInfo: &response.PerTenPageInfo{CurrPage: 2, TotalCount: 30},
	})
	refreshed := updated.(topicPage)
	require.False(t, refreshed.loading)
	require.Nil(t, refreshed.loadErr)
	require.Equal(t, 2, refreshed.page)
	require.Equal(t, 2, refreshed.index)
	require.Equal(t, int64(2), refreshed.topics[refreshed.index].Id)
	require.False(t, refreshed.refreshedAt.IsZero())
}

func TestTopicPageIgnoresStaleAndOtherNodeMessages(t *testing.T) {
	requestPageEnvironment(t)
	page := loadedRequestTopicPage()
	page.loading = true
	failure := errors.New("过期的失败")
	cases := []struct {
		name string
		msg  tea.Msg
	}{
		{"old success", messages.GetTopicResponse{NodeKey: page.node.Key, RequestID: page.requestID + 1, Data: []response.TopicResult{{Id: 99}}, PageInfo: &response.PerTenPageInfo{CurrPage: 1}}},
		{"other node success", messages.GetTopicResponse{NodeKey: "another-node", RequestID: page.requestID, Data: []response.TopicResult{{Id: 99}}, PageInfo: &response.PerTenPageInfo{CurrPage: 1}}},
		{"old failure", messages.RequestError{Kind: messages.RequestTopics, NodeKey: page.node.Key, RequestID: page.requestID + 1, Err: failure}},
		{"other node failure", messages.RequestError{Kind: messages.RequestTopics, NodeKey: "another-node", RequestID: page.requestID, Err: failure}},
		{"other kind failure", messages.RequestError{Kind: messages.RequestDetail, NodeKey: page.node.Key, RequestID: page.requestID, Err: failure}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			updated, cmd := page.Update(tc.msg)
			require.Equal(t, page, updated.(topicPage))
			require.Nil(t, cmd)
		})
	}
}

func TestDetailPageIgnoresOldRequestsEvenForSameTopic(t *testing.T) {
	requestPageEnvironment(t)
	page := loadedRequestDetailPage()
	page.loading, page.replyLoading = true, true
	page.replyRequestID = nextRequestID()
	page.decodeMap["original"] = "decoded"
	failure := errors.New("过期的失败")
	cases := []struct {
		name string
		msg  tea.Msg
	}{
		{"old detail success", messages.GetDetailResponse{TopicID: page.id, RequestID: page.requestID + 100, Data: response.V2DetailResult{Id: page.id, Title: "过期内容"}}},
		{"other topic success", messages.GetDetailResponse{TopicID: 99, RequestID: page.requestID, Data: response.V2DetailResult{Id: 99}}},
		{"old detail failure", messages.RequestError{Kind: messages.RequestDetail, TopicID: page.id, RequestID: page.requestID + 100, Err: failure}},
		{"other topic failure", messages.RequestError{Kind: messages.RequestDetail, TopicID: 99, RequestID: page.requestID, Err: failure}},
		{"old replies", messages.GetReplyResponse{TopicID: page.id, RequestID: page.replyRequestID + 100, Page: 2, Data: response.V2ReplyResponse{Result: []response.V2ReplyResult{{Id: 99}}}}},
		{"other topic replies", messages.GetReplyResponse{TopicID: 99, RequestID: page.replyRequestID, Page: 2, Data: response.V2ReplyResponse{Result: []response.V2ReplyResult{{Id: 99}}}}},
		{"old reply failure", messages.RequestError{Kind: messages.RequestReplies, TopicID: page.id, RequestID: page.replyRequestID + 100, Err: failure}},
		{"old decode", messages.DecodeDetailContentResult{TopicID: page.id, RequestID: page.requestID + 100, Result: map[string]string{"unexpected": "stale"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			updated, cmd := page.Update(tc.msg)
			actual := updated.(detailPage)
			require.Equal(t, page.contentDetail, actual.contentDetail)
			require.Equal(t, page.contentReply, actual.contentReply)
			require.Equal(t, page.replyPageInfo, actual.replyPageInfo)
			require.Equal(t, page.decodeMap, actual.decodeMap)
			require.Equal(t, page.loading, actual.loading)
			require.Equal(t, page.replyLoading, actual.replyLoading)
			require.Nil(t, actual.loadErr)
			require.Nil(t, actual.replyErr)
			require.Nil(t, cmd)
		})
	}
}

func TestReplyFailureRetriesSamePageWithoutDuplicatingComments(t *testing.T) {
	requestPageEnvironment(t)
	for _, retryKey := range []tea.KeyType{tea.KeyEnter, tea.KeyCtrlR} {
		t.Run(tea.KeyMsg{Type: retryKey}.String(), func(t *testing.T) {
			page := loadedRequestDetailPage()
			request := page.startReply(2)
			require.NotNil(t, request)
			failedID := page.replyRequestID
			failure := errors.New("评论请求超时")
			updated, _ := page.Update(messages.RequestError{Kind: messages.RequestReplies, TopicID: page.id, RequestID: failedID, Page: 2, Err: failure})
			failed := updated.(detailPage)
			require.False(t, failed.replyLoading)
			require.ErrorIs(t, failed.replyErr, failure)
			require.Equal(t, page.contentReply, failed.contentReply)
			require.Equal(t, 1, failed.replyPageInfo.CurrPage)
			require.Contains(t, failed.replyStatus(80), "重试")

			updated, retry := failed.Update(tea.KeyMsg{Type: retryKey})
			retrying := updated.(detailPage)
			require.NotNil(t, retry)
			require.True(t, retrying.replyLoading)
			require.Nil(t, retrying.replyErr)
			require.Equal(t, 2, retrying.pendingReplyPage)
			require.NotEqual(t, failedID, retrying.replyRequestID)

			// A late failure from the earlier attempt must not stop the retry.
			updated, ignored := retrying.Update(messages.RequestError{Kind: messages.RequestReplies, TopicID: page.id, RequestID: failedID, Page: 2, Err: failure})
			require.Nil(t, ignored)
			retrying = updated.(detailPage)
			require.True(t, retrying.replyLoading)
			require.Nil(t, retrying.replyErr)

			result := messages.GetReplyResponse{
				TopicID: page.id, RequestID: retrying.replyRequestID, Page: 2,
				Data: response.V2ReplyResponse{Result: []response.V2ReplyResult{{Id: 2, Content: "重叠评论"}, {Id: 3, Content: "新评论"}, {Id: 3, Content: "重复评论"}}, Pagination: response.V2PageResponse{CurrPage: 2, TotalPages: 3, TotalCount: 30}},
			}
			updated, _ = retrying.Update(result)
			loaded := updated.(detailPage)
			require.False(t, loaded.replyLoading)
			require.Nil(t, loaded.replyErr)
			require.Equal(t, 2, loaded.replyPageInfo.CurrPage)
			require.Len(t, loaded.contentReply, 3)
			require.Equal(t, "第二条评论", loaded.contentReply[1].Content)
			updated, _ = loaded.Update(result)
			require.Len(t, updated.(detailPage).contentReply, 3)
		})
	}
}

func TestDetailFailuresNeverRecordBrowsingHistory(t *testing.T) {
	for _, refreshing := range []bool{false, true} {
		name := "initial"
		if refreshing {
			name = "refresh"
		}
		t.Run(name, func(t *testing.T) {
			path := requestPageEnvironment(t)
			page := newDetailPage(42)
			var historyBefore []byte
			if refreshing {
				require.NoError(t, g.History.Record(model.HistoryEntry{TopicID: 42, Title: "旧的历史标题", VisitedAt: 123}))
				var err error
				historyBefore, err = os.ReadFile(path)
				require.NoError(t, err)
				page = loadedRequestDetailPage()
				updated, request := page.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
				require.NotNil(t, request)
				page = updated.(detailPage)
			}
			failure := errors.New("主题请求超时")
			updated, cmd := page.Update(messages.RequestError{Kind: messages.RequestDetail, TopicID: page.id, RequestID: page.requestID, Err: failure})
			failed := updated.(detailPage)
			require.Nil(t, cmd)
			require.False(t, failed.loading)
			require.ErrorIs(t, failed.loadErr, failure)
			require.Contains(t, failed.View(), "Ctrl+R")
			if refreshing {
				require.Equal(t, page.contentDetail, failed.contentDetail)
				after, err := os.ReadFile(path)
				require.NoError(t, err)
				require.Equal(t, historyBefore, after)
				require.Equal(t, int64(123), g.History.Entries()[0].VisitedAt)
			} else {
				require.Empty(t, g.History.Entries())
				_, err := os.Stat(path)
				require.ErrorIs(t, err, os.ErrNotExist)
			}
		})
	}
}

func TestEmptyDetailResponseDoesNotCreateHistory(t *testing.T) {
	path := requestPageEnvironment(t)
	page := newDetailPage(42)
	updated, cmd := page.Update(messages.GetDetailResponse{TopicID: page.id, RequestID: page.requestID})
	invalid := updated.(detailPage)
	require.False(t, invalid.loading)
	require.Error(t, invalid.loadErr)
	require.Nil(t, cmd)
	require.Empty(t, g.History.Entries())
	_, err := os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}
