package ui

import (
	"context"
	"log/slog"
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
	"github.com/seth-shi/go-v2ex/internal/ui/components/footer"
	"github.com/seth-shi/go-v2ex/internal/ui/routes"
)

type Model struct {
	contentModel tea.Model
	footerModel  tea.Model
	appVersion   string
	errConfig    error
}

func NewModel(appVersion string, errConfig error) Model {
	return Model{
		contentModel: routes.SplashModel,
		footerModel:  footer.New(appVersion),
		appVersion:   appVersion,
		errConfig:    errConfig,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		// 加载配置
		func() tea.Msg {
			slog.Info("配置加载完成", slog.Any("err", m.errConfig))
			return messages.LoadConfigResult{Error: m.errConfig}
		},
		// 其它不要用 init 初始化, 使用消息去刷新
		m.contentModel.Init(),
		m.footerModel.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msgType := msg.(type) {
	// 全局监听
	case tea.WindowSizeMsg:
		config.Screen.Width = msgType.Width
		config.Screen.Height = msgType.Height
	case messages.LoadConfigResult:
		return m, m.initHomePage(msgType.Error)
	case messages.RedirectPageRequest:
		// 切换页面
		m.contentModel = msgType.ContentModel
		return m, tea.Sequence(messages.Post(messages.ShowAlertRequest{Text: ""}))
	case messages.GetTopicResponse:
		// 缓存这个列表, 进到详情页回来还有数据, 并且消息传递给子级
		routes.TopicsModel.SetTopics(msgType)
	case messages.RedirectDetailRequest:
		return m, tea.Sequence(
			messages.Post(messages.RedirectPageRequest{ContentModel: routes.DetailModel}),
			messages.Post(messages.GetDetailRequest{ID: msgType.Id}),
		)
	case messages.RedirectTopicsPage:
		var cmds = []tea.Cmd{
			messages.Post(messages.RedirectPageRequest{ContentModel: routes.TopicsModel}),
		}
		if msgType.Page > 0 {
			cmds = append(cmds, messages.Post(messages.GetTopicsRequest{Page: msgType.Page}))
		}
		return m, tea.Sequence(cmds...)
	case tea.KeyMsg:
		switch {
		case key.Matches(msgType, consts.AppKeyMap.Space):
			config.Session.BossComingMode = !config.Session.BossComingMode
			return m, m.returnPage(routes.BossComingModel)
		case key.Matches(msgType, consts.AppKeyMap.SettingPage):
			return m, tea.Sequence(
				m.returnPage(routes.SettingModel),
				messages.Post(messages.RefreshSettingInputRequest{}),
			)
		case key.Matches(msgType, consts.AppKeyMap.HelpPage):
			return m, m.returnPage(routes.HelpModel)
		case key.Matches(msgType, consts.AppKeyMap.SwitchShowMode):
			config.G.SwitchShowMode()
			return m, tea.Batch(
				messages.ErrorOrToast(config.SaveToFile, ""),
				messages.Post(messages.ShowToastRequest{Text: config.G.GetShowModeText()}),
			)
		case key.Matches(msgType, consts.AppKeyMap.Quit):
			return m, tea.Quit
		}
	}

	// 更新当前的主要三个部分组件
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)
	m.contentModel, cmd = m.contentModel.Update(msg)
	cmds = append(cmds, cmd)
	m.footerModel, cmd = m.footerModel.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) returnPage(contentModel tea.Model) tea.Cmd {
	if reflect.DeepEqual(m.contentModel, contentModel) {
		return m.initHomePage(nil)
	}

	return messages.Post(messages.RedirectPageRequest{ContentModel: contentModel})
}

func (m Model) View() string {

	var (
		output strings.Builder
	)

	output.WriteString(m.contentModel.View())
	// 底部增加一个 padding, 来固定在底部
	if !config.Session.BossComingMode {
		ff := m.footerModel.View()
		// 打印调试信息
		paddingTop := config.Screen.Height -
			lipgloss.Height(output.String()) -
			lipgloss.Height(ff)
		output.WriteString("\n")
		output.WriteString(lipgloss.NewStyle().PaddingTop(paddingTop).Render(ff))
	}

	return output.String()
}

func (m Model) initHomePage(err error) tea.Cmd {

	// 把配置注入到其他页面
	var cmds = []tea.Cmd{
		// 读取配置文件有错误, 不影响后续流程, 可以让用户自己抉择
		messages.Post(err),
	}

	// 没 token 去配置页面
	if config.G.Token == "" {
		cmds = append(
			cmds,
			messages.Post(messages.RedirectPageRequest{ContentModel: routes.SettingModel}),
			messages.Post(messages.ShowToastRequest{Text: "请先按照说明配置秘钥和节点"}),
		)
		return tea.Sequence(cmds...)
	}

	// 去触发对应的地方获取数据
	cmds = append(
		cmds,
		// 先跳转到主题页, 然后获取第一页的数据
		tea.Sequence(
			messages.Post(messages.RedirectPageRequest{ContentModel: routes.TopicsModel}),
			messages.Post(messages.GetTopicsRequest{Page: 1}),
		),
		// 获取个人信息
		tea.Sequence(
			messages.LoadingGetToken.PostStart(),
			api.V2ex.GetToken(context.Background()),
			messages.LoadingGetToken.PostEnd(),
		),
	)
	return tea.Sequence(cmds...)
}
