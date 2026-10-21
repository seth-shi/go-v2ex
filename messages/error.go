package messages

import (
	tea "github.com/charmbracelet/bubbletea"
)

type RequestKind string

const (
	RequestTopics  RequestKind = "topics"
	RequestDetail  RequestKind = "detail"
	RequestReplies RequestKind = "replies"
)

// RequestError keeps failures tied to the page and request that produced them.
// Pages can ignore stale failures without mistaking them for a new request.
type RequestError struct {
	Kind      RequestKind
	NodeKey   string
	TopicID   int64
	RequestID uint64
	Page      int
	Err       error
}

func (e RequestError) Error() string {
	if e.Err == nil {
		return "请求失败"
	}
	return e.Err.Error()
}

func (e RequestError) Unwrap() error { return e.Err }

func ErrorOrToast(err error, msg string) tea.Msg {

	if err != nil {
		return err
	}

	if msg == "" {
		return nil
	}

	return ProxyShowToastRequest{Text: msg}

}
