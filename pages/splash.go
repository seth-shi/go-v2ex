package pages

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/api"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/pkg"
)

type splashPage struct {
}

func newSplashPage() splashPage {
	return splashPage{}
}

func (m splashPage) Init() tea.Cmd {

	return tea.Batch(
		commands.LoadConfig(),
	)
}

func (m splashPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case messages.LoadConfigResult:
		return m.onConfigResult(msg)
	}

	return m, nil
}

func (m splashPage) onConfigResult(msg messages.LoadConfigResult) (tea.Model, tea.Cmd) {

	conf := msg.Result
	pkg.SetupLogger(conf)
	api.SetUpHttpClient(conf)
	pkg.SetUpImageHttpClient(conf)

	// 把配置注入到其他页面
	var cmds = []tea.Cmd{
		// 检查版本更新
		commands.AlertError(msg.Err),
		commands.Post(messages.CheckUpgradeAppRequest{}),
	}

	// 没 token 去配置页面
	if conf.Token == "" {
		cmds = append(
			cmds,
			commands.Redirect(RouteSetting),
			commands.AlertInfo("请先按照说明配置秘钥和节点"),
		)
		return m, tea.Sequence(cmds...)
	}

	// 去触发对应的地方获取数据
	cmds = append(
		cmds,
		// 获取 token 过期信息
		tea.Sequence(
			messages.LoadingGetToken.PostStart(),
			api.V2ex.GetToken(context.Background()),
			messages.LoadingGetToken.PostEnd(),
		),
		// 获取个人信息
		tea.Sequence(
			messages.LoadingMe.PostStart(),
			api.V2ex.Me(context.Background()),
			messages.LoadingMe.PostEnd(),
		),
		// 先跳转到主题页
		commands.Redirect(RouteTopic),
	)
	return m, tea.Sequence(cmds...)
}

func (m splashPage) View() string {
	return loadingView("开屏页...")
}
