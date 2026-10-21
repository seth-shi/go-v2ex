package nav

import (
	"errors"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/messages"
)

type responseLifecycle struct {
	entered int
	left    int
}

type responsePage struct {
	requestID uint64
	topicID   int64
	updates   int
	last      tea.Msg
	lifecycle *responseLifecycle
}

func (m responsePage) Init() tea.Cmd { return nil }
func (m responsePage) View() string  { return "page" }
func (m responsePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updates++
	m.last = msg
	return m, func() tea.Msg { return responseEffect(m.updates) }
}
func (m responsePage) OnEntering() (tea.Model, tea.Cmd) {
	m.lifecycle.entered++
	return m, nil
}
func (m responsePage) OnLeaving() (tea.Model, tea.Cmd) {
	m.lifecycle.left++
	return m, nil
}
func (m responsePage) AcceptsMessage(msg tea.Msg) bool {
	if m.requestID == 0 {
		return false
	}
	switch msg := msg.(type) {
	case messages.GetTopicResponse:
		return msg.RequestID == m.requestID
	case messages.GetDetailResponse:
		return msg.RequestID == m.requestID && msg.TopicID == m.topicID
	case messages.GetReplyResponse:
		return msg.RequestID == m.requestID && msg.TopicID == m.topicID
	case messages.DecodeDetailContentResult:
		return msg.RequestID == m.requestID && msg.TopicID == m.topicID
	case messages.RequestError:
		return msg.RequestID == m.requestID && msg.TopicID == m.topicID
	}
	return false
}

type responseEffect int

func installResponseNavigator(t *testing.T, current tea.Model, histories ...tea.Model) {
	t.Helper()
	previous := nav
	nav = &navigator{curr: current, histories: histories}
	t.Cleanup(func() { nav = previous })
}

func TestScopedResponsesUpdateHiddenOwnerWithoutChangingCurrentPage(t *testing.T) {
	cases := map[string]tea.Msg{
		"topics":  messages.GetTopicResponse{RequestID: 7},
		"detail":  messages.GetDetailResponse{RequestID: 7, TopicID: 42},
		"replies": messages.GetReplyResponse{RequestID: 7, TopicID: 42},
		"decode":  messages.DecodeDetailContentResult{RequestID: 7, TopicID: 42},
		"failure": messages.RequestError{RequestID: 7, TopicID: 42, Kind: messages.RequestDetail, Err: errors.New("timeout")},
	}
	for name, msg := range cases {
		t.Run(name, func(t *testing.T) {
			lifecycle := &responseLifecycle{}
			installResponseNavigator(t, responsePage{}, responsePage{requestID: 7, topicID: 42, lifecycle: lifecycle})

			cmd := Update(msg)
			if cmd == nil || cmd() != responseEffect(1) {
				t.Fatal("hidden page's update command was lost")
			}
			if current := nav.curr.(responsePage); current.updates != 0 {
				t.Fatal("response was sent to the covering page")
			}
			if saved := nav.histories[0].(responsePage); saved.updates != 1 || !reflect.DeepEqual(saved.last, msg) {
				t.Fatal("updated hidden page was not saved in the navigation stack")
			}
			if lifecycle.entered != 0 || lifecycle.left != 0 {
				t.Fatal("routing a response must not invoke page lifecycle hooks")
			}
		})
	}
}

func TestScopedResponsesPreferCurrentOwner(t *testing.T) {
	installResponseNavigator(t, responsePage{requestID: 7, topicID: 42}, responsePage{requestID: 7, topicID: 42})
	Update(messages.GetDetailResponse{RequestID: 7, TopicID: 42})
	if nav.curr.(responsePage).updates != 1 || nav.histories[0].(responsePage).updates != 0 {
		t.Fatal("current page should receive its response before a saved matching page")
	}
}

func TestScopedResponsesPreferMostRecentHiddenOwner(t *testing.T) {
	installResponseNavigator(t, responsePage{}, responsePage{requestID: 7, topicID: 42}, responsePage{requestID: 7, topicID: 42})
	Update(messages.GetReplyResponse{RequestID: 7, TopicID: 42})
	if nav.histories[0].(responsePage).updates != 0 || nav.histories[1].(responsePage).updates != 1 {
		t.Fatal("only the most recent matching hidden page should be updated")
	}
}

func TestScopedResponsesDropExpiredRequestsAndOtherTopics(t *testing.T) {
	for _, msg := range []tea.Msg{
		messages.GetDetailResponse{RequestID: 6, TopicID: 42},
		messages.GetDetailResponse{RequestID: 7, TopicID: 99},
		messages.RequestError{RequestID: 6, TopicID: 42, Kind: messages.RequestDetail, Err: errors.New("stale timeout")},
	} {
		installResponseNavigator(t, responsePage{}, responsePage{requestID: 7, topicID: 42})
		if Update(msg) != nil {
			t.Fatal("unmatched response should be discarded")
		}
		if nav.curr.(responsePage).updates != 0 || nav.histories[0].(responsePage).updates != 0 {
			t.Fatal("unmatched response changed a page")
		}
	}
}

func TestOrdinaryInputOnlyUpdatesCurrentPage(t *testing.T) {
	installResponseNavigator(t, responsePage{}, responsePage{requestID: 7, topicID: 42})
	Update(tea.KeyMsg{Type: tea.KeyDown})
	if nav.curr.(responsePage).updates != 1 || nav.histories[0].(responsePage).updates != 0 {
		t.Fatal("keyboard input must not be routed to hidden pages")
	}
}
