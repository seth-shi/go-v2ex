package pages

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/v2/api"
	"github.com/seth-shi/go-v2ex/v2/commands"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/messages"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/nav"
	"github.com/seth-shi/go-v2ex/v2/pkg"
)

var _ nav.PageLife = splashPage{}

type splashPage struct {
	showWelcome bool
	continued   bool
}

func newSplashPage() splashPage {
	return splashPage{}
}

func (m splashPage) Init() tea.Cmd {
	return tea.Batch(
		commands.LoadConfig(),
	)
}

func (m splashPage) OnEntering() (tea.Model, tea.Cmd) {
	g.Session.HideFooter.Store(m.showWelcome)
	return m, nil
}

func (m splashPage) OnLeaving() (tea.Model, tea.Cmd) {
	g.Session.HideFooter.Store(false)
	return m, nil
}

func (m splashPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case messages.LoadConfigResult:
		return m.onConfigResult(msg)
	case tea.KeyMsg:
		if m.showWelcome && msg.String() == "enter" {
			return m.continueAfterWelcome()
		}
	}

	return m, nil
}

func (m splashPage) onConfigResult(msg messages.LoadConfigResult) (tea.Model, tea.Cmd) {

	// 初始化日志, HTTP 请求
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
	if !conf.OnboardingDone {
		m.showWelcome = true
		g.Session.HideFooter.Store(true)
		return m, tea.Sequence(cmds...)
	}

	m.continued = true
	cmds = append(cmds, mainPageCommands(conf)...)
	return m, tea.Sequence(cmds...)
}

func (m splashPage) continueAfterWelcome() (tea.Model, tea.Cmd) {
	if m.continued {
		return m, nil
	}

	m.showWelcome = false
	m.continued = true
	g.Session.HideFooter.Store(false)
	save := func() tea.Msg {
		err := g.Config.Save(func(conf *model.FileConfig) {
			conf.OnboardingDone = true
		})
		return messages.ErrorOrToast(err, "欢迎使用 Go V2EX")
	}
	cmds := []tea.Cmd{save}
	cmds = append(cmds, mainPageCommands(g.Config.Get())...)
	return m, tea.Sequence(cmds...)
}

func mainPageCommands(conf *model.FileConfig) []tea.Cmd {
	cmds := []tea.Cmd{nav.Push(newTopicPage())}
	if strings.TrimSpace(conf.Token) == "" {
		return cmds
	}

	// 先显示主题页，再在后台检查令牌和个人信息。
	return append(cmds, tea.Batch(
		commands.LoadingGetToken.Run(api.V2ex.GetToken(context.Background())),
		commands.LoadingMe.Run(api.V2ex.Me(context.Background())),
	))
}

func (m splashPage) View() string {
	if m.showWelcome {
		return welcomeView()
	}
	return loadingView("正在连接 V2EX")
}
