package messages

import (
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	lastLoadingId        int64
	LoadingGetToken      = newLoadingKey("获取 token 信息中")
	LoadingRequestTopics = newLoadingKey("获取主题中")
	LoadingRequestDetail = newLoadingKey("获取内容中")
	LoadingRequestReply  = newLoadingKey("获取评论中")
	LoadingRequestImage  = newLoadingKey("获取图片中")
)

func newLoadingKey(text string) loadingManager {
	id := nextLoadingId()
	return loadingManager{
		start: StartLoading{Text: text, ID: id},
		end: EndLoading{
			ID: id,
		},
		loading: atomic.Bool{},
	}
}

type StartLoading struct {
	Text string
	ID   int
}

type EndLoading struct {
	ID int
}

func (s *loadingManager) PostStart() tea.Cmd {

	if s.loading.Load() {
		return nil
	}

	return func() tea.Msg {
		s.loading.Store(true)
		return s.start
	}
}

func (s *loadingManager) Loading() bool {
	return s.loading.Load()
}

func (s *loadingManager) PostEnd() tea.Cmd {
	return func() tea.Msg {
		s.loading.Store(false)
		return s.end
	}
}

type loadingManager struct {
	start   StartLoading
	end     EndLoading
	loading atomic.Bool
}

func nextLoadingId() int {
	return int(atomic.AddInt64(&lastLoadingId, 1))
}
