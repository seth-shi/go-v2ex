package commands

import (
	"fmt"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/messages"
)

var (
	lastLoadingId        int64
	LoadingGetToken      = newLoadingKey("token 获取中...")
	LoadingMe            = newLoadingKey("个人信息获取中...")
	LoadingRequestTopics = newLoadingKey("主题请求中...")
	LoadingRequestDetail = newLoadingKey("内容获取中...")
	LoadingRequestReply  = newLoadingKey("评论获取中...")
	LoadingDecodeContent = newLoadingKey("内容解码中...")
)

type loadingManager struct {
	start   messages.StartLoading
	end     messages.EndLoading
	loading atomic.Bool
}

func newLoadingKey(text string) loadingManager {
	id := nextLoadingId()
	return loadingManager{
		start: messages.StartLoading{Text: text, ID: id},
		end: messages.EndLoading{
			ID: id,
		},
		loading: atomic.Bool{},
	}
}

func (s *loadingManager) Run(cmd tea.Cmd) tea.Cmd {

	// 加载中的时候, 发出错误
	if s.loading.Load() {
		return Post(fmt.Errorf("[%s]加载中", s.start.Text))
	}

	// 开始加载中
	return tea.Sequence(
		// 标记开始
		func() tea.Msg {
			s.loading.Store(true)
			return s.start
		},
		cmd,
		// 标记结束
		func() tea.Msg {
			s.loading.Store(false)
			return s.end
		},
	)
}

func nextLoadingId() int {
	return int(atomic.AddInt64(&lastLoadingId, 1))
}
