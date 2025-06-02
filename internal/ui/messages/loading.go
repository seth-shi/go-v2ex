package messages

import (
	"sync/atomic"
)

var (
	lastLoadingId        int64
	LoadingGetToken      = newLoadingKey("获取 token 信息中")
	LoadingRequestTopics = newLoadingKey("获取主题中")
	LoadingRequestDetail = newLoadingKey("获取内容中")
	LoadingRequestReply  = newLoadingKey("获取评论中")
)

func newLoadingKey(text string) loadingCombine {
	id := nextLoadingId()
	return loadingCombine{
		Start: StartLoading{
			ID:   id,
			Text: text,
		},
		End: EndLoading{
			ID: id,
		},
	}
}

type StartLoading struct {
	Text string
	ID   int
}

type EndLoading struct {
	ID int
}

func nextLoadingId() int {
	return int(atomic.AddInt64(&lastLoadingId, 1))
}

type loadingCombine struct {
	Start StartLoading
	End   EndLoading
}
