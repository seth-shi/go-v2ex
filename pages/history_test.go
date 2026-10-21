package pages

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/stretchr/testify/require"
)

func useTemporaryHistory(t *testing.T) string {
	t.Helper()
	originalHistory, originalWindow := g.History, g.Window
	originalConfig, originalSession := g.Config, g.Session
	t.Cleanup(func() {
		g.History, g.Window = originalHistory, originalWindow
		g.Config, g.Session = originalConfig, originalSession
	})
	historyPath := filepath.Join(t.TempDir(), "history.json")
	g.History = model.NewHistoryStore(historyPath)
	g.Window = &model.WindowSize{}
	g.Window.SetSize(tea.WindowSizeMsg{Width: 60, Height: 28})
	g.Config = model.NewSafe(model.NewDefaultFileConfig(), nil)
	g.Session = &model.SessionData{}
	return historyPath
}

// Detail success currently batches a history write followed by a status
// update. With Replies=0 there is no reply request. Execute only the first,
// explicitly known history command, never arbitrary page/network commands.
func recordSuccessfulDetail(t *testing.T, id int64, title string) {
	t.Helper()
	page := newDetailPage(id)
	entered, _ := page.OnEntering()
	page = entered.(detailPage)
	updated, command := page.Update(messages.GetDetailResponse{
		TopicID:   id,
		RequestID: page.requestID,
		Data: response.V2DetailResult{
			Id:      id,
			Title:   title,
			Content: "已成功获取的正文",
			Url:     fmt.Sprintf("https://www.v2ex.com/t/%d", id),
			Replies: 0,
			Node:    response.NodeInfoResult{Title: "程序员"},
			Member:  response.MemberResult{Id: 1, Username: "alice"},
		},
	})
	require.False(t, updated.(detailPage).loading)
	require.NoError(t, updated.(detailPage).loadErr)
	require.NotNil(t, command)
	batch, ok := command().(tea.BatchMsg)
	require.True(t, ok, "detail success should batch its history and status commands")
	require.Len(t, batch, 2)
	result, ok := batch[0]().(historyUpdated)
	require.True(t, ok, "the first success command must report the history write")
	require.NoError(t, result.Err)
}

func TestSuccessfulDetailPersistsHistoryAndReopensAfterRestart(t *testing.T) {
	historyPath := useTemporaryHistory(t)
	recordSuccessfulDetail(t, 41, "原先标题")
	recordSuccessfulDetail(t, 42, "另一篇主题")
	recordSuccessfulDetail(t, 41, "\x1b[31m更新标题\x1b[0m\n继续")

	entries := g.History.Entries()
	require.Len(t, entries, 2, "revisiting a topic must not create another history row")
	require.EqualValues(t, 41, entries[0].TopicID)
	require.Equal(t, "更新标题 继续", entries[0].Title)
	require.Equal(t, "程序员", entries[0].Node)
	require.Equal(t, "alice", entries[0].Author)
	require.Equal(t, "https://www.v2ex.com/t/41", entries[0].URL)
	require.Positive(t, entries[0].VisitedAt)

	// Replace the global store to simulate a new application session. Opening
	// history must load from disk, not rely on the previous store's memory.
	g.History = model.NewHistoryStore(historyPath)
	page := newHistoryPage()
	entered, command := page.OnEntering()
	page = entered.(historyPage)
	require.Empty(t, page.entries)
	require.NotNil(t, command)
	batch, ok := command().(tea.BatchMsg)
	require.True(t, ok)
	require.Len(t, batch, 2)
	loaded, ok := batch[0]().(historyUpdated)
	require.True(t, ok)
	require.NoError(t, loaded.Err)
	updated, _ := page.Update(loaded)
	page = updated.(historyPage)
	require.True(t, page.loaded)
	require.Len(t, page.entries, 2)
	require.EqualValues(t, 41, page.entries[0].TopicID)
	require.Contains(t, ansi.Strip(page.View()), "更新标题 继续")

	// Check both the recent entry and another selected row. Inspect the push
	// message without applying navigation, which would initialize network IO.
	for _, wantID := range []int64{41, 42} {
		_, open := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
		require.NotNil(t, open)
		target, ok := open().(detailPage)
		require.True(t, ok, "Enter must navigate to a detail page")
		require.Equal(t, wantID, target.id)
		moved, _ := page.Update(tea.KeyMsg{Type: tea.KeyDown})
		page = moved.(historyPage)
	}
}

func TestFailedStaleAndMismatchedDetailsDoNotRecordHistory(t *testing.T) {
	historyPath := useTemporaryHistory(t)
	for _, scenario := range []string{"failed", "stale", "wrong topic", "invalid detail"} {
		t.Run(scenario, func(t *testing.T) {
			page := newDetailPage(41)
			var message tea.Msg = messages.GetDetailResponse{
				TopicID:   41,
				RequestID: page.requestID,
				Data:      response.V2DetailResult{Id: 41, Title: "unrecorded"},
			}
			switch scenario {
			case "failed":
				message = messages.RequestError{Kind: messages.RequestDetail, TopicID: 41, RequestID: page.requestID, Err: errors.New("offline")}
			case "stale":
				response := message.(messages.GetDetailResponse)
				response.RequestID = page.requestID + 1
				message = response
			case "wrong topic":
				response := message.(messages.GetDetailResponse)
				response.TopicID = 42
				message = response
			case "invalid detail":
				response := message.(messages.GetDetailResponse)
				response.Data.Id = 0
				message = response
			}
			_, command := page.Update(message)
			require.Nil(t, command, "an unsuccessful response must not schedule a history write")
			require.Empty(t, g.History.Entries())
			_, err := os.Stat(historyPath)
			require.True(t, os.IsNotExist(err), "unsuccessful requests must not create history")
		})
	}
}

func TestEmptyHistoryCannotOpenTopic(t *testing.T) {
	useTemporaryHistory(t)
	page := newHistoryPage()
	updated, _ := page.Update(loadHistory()())
	page = updated.(historyPage)
	require.Contains(t, page.View(), "还没有阅读历史")
	for _, key := range []tea.KeyMsg{{Type: tea.KeyEnter}, {Type: tea.KeyUp}, {Type: tea.KeyDown}} {
		updated, command := page.Update(key)
		require.Nil(t, command)
		page = updated.(historyPage)
		require.Zero(t, page.index)
	}
}

func TestHistoryListKeepsLastSelectedEntryVisibleAtCommonSizes(t *testing.T) {
	useTemporaryHistory(t)
	entries := make([]model.HistoryEntry, 200)
	for i := range entries {
		entries[i] = model.HistoryEntry{
			TopicID:   int64(i + 1),
			Title:     fmt.Sprintf("记录%03d 中文长标题用于检查末项及截断效果", i+1),
			Node:      "程序员",
			Author:    fmt.Sprintf("author%03d", i+1),
			VisitedAt: 1_756_092_800,
		}
	}
	for _, size := range []tea.WindowSizeMsg{
		{Width: 32, Height: 12},
		{Width: 40, Height: 24},
		{Width: 60, Height: 28},
		{Width: 100, Height: 32},
	} {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			g.Window.SetSize(size)
			page := historyPage{entries: entries, loaded: true}
			view := ansi.Strip(page.View())
			assertViewFits(t, view, size.Width, size.Height-1)
			require.Contains(t, view, "记录001")
			require.Contains(t, view, "author001")

			// Up from the first row wraps to the final row, forcing the list
			// viewport to reveal its title and metadata above the footer.
			updated, command := page.Update(tea.KeyMsg{Type: tea.KeyUp})
			require.Nil(t, command)
			page = updated.(historyPage)
			require.Equal(t, 199, page.index)
			view = ansi.Strip(page.View())
			assertViewFits(t, view, size.Width, size.Height-1)
			require.Contains(t, view, "记录200")
			require.Contains(t, view, "author200", "the last selected row's metadata must not be clipped")
			require.NotContains(t, view, "记录001")
		})
	}
}
