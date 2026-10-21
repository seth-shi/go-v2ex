package nav

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/messages"
)

var (
	nav      = &navigator{}
	maxPages = 1000
)

type navigator struct {
	histories []tea.Model
	curr      tea.Model
}

// MessageReceiver identifies the page that owns an asynchronous response.
// Implementations must match the request ID as well as the resource identity,
// so older responses cannot overwrite a newer request or another page.
type MessageReceiver interface {
	AcceptsMessage(tea.Msg) bool
}

func View() string {
	if nav.curr == nil {
		return ""
	}

	return nav.curr.View()
}

func Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case PushNavigationMsg:
		return handlePushNavigation(msg)
	case PopNavigationMsg:
		return handlePopNavigation()
	case messages.GetTopicResponse, messages.GetDetailResponse,
		messages.GetReplyResponse, messages.DecodeDetailContentResult,
		messages.RequestError:
		return routeResponse(msg)
	}

	if nav.curr == nil {
		return nil
	}

	// update page
	var c tea.Cmd
	nav.curr, c = nav.curr.Update(msg)
	return c
}

func routeResponse(msg tea.Msg) tea.Cmd {
	if receiver, ok := nav.curr.(MessageReceiver); ok && receiver.AcceptsMessage(msg) {
		var cmd tea.Cmd
		nav.curr, cmd = nav.curr.Update(msg)
		return cmd
	}

	// A request keeps running when help, settings or history covers its page.
	// Update the saved model without entering it or disturbing the active page.
	for i := len(nav.histories) - 1; i >= 0; i-- {
		page := nav.histories[i]
		if receiver, ok := page.(MessageReceiver); ok && receiver.AcceptsMessage(msg) {
			updated, cmd := page.Update(msg)
			nav.histories[i] = updated
			return cmd
		}
	}

	return nil
}

// handlePushNavigation 处理页面压入导航栈的操作
func handlePushNavigation(msg PushNavigationMsg) tea.Cmd {
	var (
		cmds []tea.Cmd
		c    tea.Cmd
	)

	// 1. 处理当前页面的离开逻辑
	if nav.curr != nil {
		// 将当前页面存入历史记录
		nav.histories = append(nav.histories, nav.curr)
		if len(nav.histories) > maxPages {
			first := nav.histories[0]
			if p, ok := first.(PageLife); ok {
				_, c = p.OnLeaving()
				cmds = append(cmds, c)
			}

			nav.histories = nav.histories[1:]
		}

		// 调用当前页面的离开钩子
		if p, ok := nav.curr.(PageLife); ok {
			nav.curr, c = p.OnLeaving()
			cmds = append(cmds, c)
		}
	}

	// 2. 压入新页面
	nav.curr = tea.Model(msg)
	// 3. 初始化新页面并处理进入逻辑
	cmds = append(cmds, nav.curr.Init())

	// 页面离开钩子
	if p, ok := nav.curr.(PageLife); ok {
		nav.curr, c = p.OnEntering()
		cmds = append(cmds, c)
	}

	return tea.Sequence(cmds...)
}

// handlePopNavigation 处理页面弹出导航栈的操作
func handlePopNavigation() tea.Cmd {

	// 尝试弹出页面
	if nav.curr == nil {
		return postCmd(NoMorePageMsg{})
	}

	if len(nav.histories) == 0 {
		return postCmd(NoMorePageMsg{})
	}

	var (
		cmds []tea.Cmd
		c    tea.Cmd
	)
	// 1. 处理旧页面的离开逻辑
	if p, ok := nav.curr.(PageLife); ok {
		_, c = p.OnLeaving()
		cmds = append(cmds, c)
	}

	// 2. 处理新页面的进入逻辑
	nav.curr = nav.histories[len(nav.histories)-1]
	nav.histories = nav.histories[:len(nav.histories)-1]

	if p, ok := nav.curr.(PageLife); ok {
		m, cmd := p.OnEntering()
		nav.curr = m
		cmds = append(cmds, cmd)
	}

	return tea.Sequence(cmds...)
}
