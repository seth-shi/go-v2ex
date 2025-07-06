package pages

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/commands"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/pkg"
)

type splashPage struct {
	windowPage
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

	m.windowPage = m.windowPage.Update(msg)

	switch msg := msg.(type) {
	case messages.LoadConfigResult:
		return m.onConfigResult(msg.Result)
	}

	return m, nil
}

func (m splashPage) onConfigResult(conf *config.FileConfig) (tea.Model, tea.Cmd) {

	pkg.SetupLogger(conf)
	api.SetUpHttpClient(conf)
	pkg.SetUpImageHttpClient(conf)
	// return m, messages.RedirectCmd(RouteHelp)

	// 把配置注入到其他页面
	var cmds = []tea.Cmd{
		// 检查版本更新
		messages.Post(messages.CheckUpgradeAppRequest{}),
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
		// 先跳转到主题页, 然后获取第一页的数据
		tea.Sequence(
			commands.Redirect(RouteTopic),
			commands.Post(messages.GetTopicsRequest{Page: 1}),
		),
		// 获取个人信息
		tea.Sequence(
			messages.LoadingGetToken.PostStart(),
			api.V2ex.GetToken(context.Background()),
			messages.LoadingGetToken.PostEnd(),
		),
	)
	return m, tea.Sequence(cmds...)
}

func (m splashPage) View() string {
	return loadingView(m.w, m.h, "开屏页...")
}
