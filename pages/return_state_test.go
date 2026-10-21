package pages

import (
	"testing"

	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/stretchr/testify/require"
)

func TestReturningFromSettingsReloadsChangedCustomNodes(t *testing.T) {
	requestPageEnvironment(t)
	conf := *g.Config.Get()
	conf.ActiveTab = 10
	conf.MyNodes = "programmer"
	g.Config.Set(&conf)
	page := newTopicPage()
	page.loading = false
	page.topics = []response.TopicResult{{Id: 1, Title: "旧节点的主题"}}
	changed := conf
	changed.MyNodes = "qna,create"
	g.Config.Set(&changed)
	updated, cmd := page.OnEntering()
	actual := updated.(topicPage)
	require.NotNil(t, cmd)
	require.True(t, actual.loading)
	require.True(t, actual.fresh)
	require.Equal(t, []string{"qna", "create"}, actual.node.Nodes)
	require.Empty(t, actual.topics)
	require.NotEqual(t, page.requestID, actual.requestID)
}

func TestReturningToHistoryKeepsTheSameTopicAfterReordering(t *testing.T) {
	requestPageEnvironment(t)
	require.NoError(t, g.History.Record(model.HistoryEntry{TopicID: 1, Title: "one"}))
	require.NoError(t, g.History.Record(model.HistoryEntry{TopicID: 2, Title: "two"}))
	page := historyPage{entries: g.History.Entries(), index: 1, loaded: true}
	require.NoError(t, g.History.Record(model.HistoryEntry{TopicID: 1, Title: "one"}))
	updated, _ := page.OnEntering()
	actual := updated.(historyPage)
	require.Equal(t, int64(1), actual.entries[actual.index].TopicID)
	require.Equal(t, 0, actual.index)
}

func TestReturningToDetailRestoresItsStatus(t *testing.T) {
	requestPageEnvironment(t)
	page := loadedRequestDetailPage()
	_, cmd := page.OnEntering()
	require.NotNil(t, cmd)
	status := cmd().(messages.ShowStatusBarTextRequest)
	require.Equal(t, "detailPage", status.Page)
	require.NotEmpty(t, status.FirstText)
}

func TestRefreshFollowsLastPageWhenTheListShrinks(t *testing.T) {
	requestPageEnvironment(t)
	page := newTopicPage()
	page.page = 3
	page.pendingPage = 3
	page.fresh = true
	page.topics = []response.TopicResult{{Id: 42, Title: "刷新前的主题"}}
	updated, cmd := page.Update(messages.GetTopicResponse{RequestID: page.requestID, NodeKey: page.node.Key, PageInfo: &response.PerTenPageInfo{CurrPage: 3, TotalCount: 15}})
	actual := updated.(topicPage)
	require.NotNil(t, cmd)
	require.Equal(t, 2, actual.pendingPage)
	require.True(t, actual.loading)
	require.Equal(t, page.topics, actual.topics)
	require.NotEqual(t, page.requestID, actual.requestID)
}
