package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevm/bubbleo/navstack"
	"github.com/seth-shi/go-v2ex/internal/commands"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
)

var (
	RouteSplash = navstack.NavigationItem{
		Title: "开屏页",
		Model: newSplashPage(),
	}
	RouteHelp = navstack.NavigationItem{
		Title: "帮助页",
		Model: newHelpPage(),
	}
	RouteBoss = navstack.NavigationItem{
		Title: "boss",
		Model: newBossPage(),
	}
	RouteSetting = navstack.NavigationItem{
		Title: "设置页",
		Model: newSettingPage(),
	}
	RouteTopic = navstack.NavigationItem{
		Title: "首页",
		Model: newTopicPage(),
	}
	RouteDetail = navstack.NavigationItem{
		Title: "详情页",
		Model: newDetailPage(),
	}
)

// 如果当前页相等, 那么返回上一页, 否则跳转过去
func redirectIfSamePop(top *navstack.NavigationItem, item navstack.NavigationItem) tea.Cmd {

	// 如果当前无东西, 直接返回
	var cmds []tea.Cmd
	// 所有页面都默认展示底部栏, 如果不想展示, 在特定页面 Init 方法隐藏
	cmds = append(cmds, messages.Post(messages.FooterStatusMessage{HiddenFooter: false}))

	if top != nil && top.Title == item.Title {
		cmds = append(cmds, commands.RedirectPop())
	} else {
		cmds = append(cmds, commands.Redirect(item))
	}

	return tea.Sequence(cmds...)
}
