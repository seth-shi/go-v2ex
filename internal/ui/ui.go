package ui

import (
	"reflect"
	"strings"

	"github.com/seth-shi/go-v2ex/internal/config"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/ui/components/footer"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
	"github.com/seth-shi/go-v2ex/internal/ui/routes"
)

type Model struct {
	contentModel tea.Model
	footerModel  tea.Model
}

func NewModel() Model {
	return Model{
		contentModel: routes.SplashModel,
		footerModel:  footer.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		// 加载配置
		config.LoadFileConfig,
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
		return m, tea.Sequence(messages.Post(messages.ShowTipsRequest{Text: ""}))
	case messages.GetTopicsResult:
		// 缓存这个列表, 进到详情页回来还有数据, 并且消息传递给子级
		routes.TopicsModel.SetTopics(msgType.Topics)
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
		case key.Matches(msgType, consts.AppKeyMap.SettingPage):
			if reflect.DeepEqual(m.contentModel, routes.SettingModel) {
				return m, m.initHomePage(nil)
			}
			return m, messages.Post(messages.RedirectPageRequest{ContentModel: routes.SettingModel})
		case key.Matches(msgType, consts.AppKeyMap.HelpPage):
			if reflect.DeepEqual(m.contentModel, routes.HelpModel) {
				return m, m.initHomePage(nil)
			}
			return m, messages.Post(messages.RedirectPageRequest{ContentModel: routes.HelpModel})
		case key.Matches(msgType, consts.AppKeyMap.SwitchShowMode):
			if !reflect.DeepEqual(m.contentModel, routes.SettingModel) {
				config.G.SwitchShowMode()
				return m, config.SaveToFile("")
			}
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

func (m Model) View() string {

	var (
		output strings.Builder
	)

	output.WriteString(m.contentModel.View())
	output.WriteRune('\n')

	// 底部增加一个 padding, 来固定在底部
	ff := m.footerModel.View()
	paddingTop := config.Screen.Height - lipgloss.Height(output.String()) - lipgloss.Height(ff)
	output.WriteString(lipgloss.NewStyle().PaddingTop(paddingTop).Render(ff))

	return output.String()
}

func (m Model) initHomePage(err error) tea.Cmd {

	// 把配置注入到其他页面
	api.Client.RefreshConfig()
	routes.SettingModel.RefreshConfig()

	var cmds = []tea.Cmd{
		// 读取配置文件有错误, 不影响后续流程, 可以让用户自己抉择
		messages.Post(err),
		//  LaunchesCount ++ 了, 保存配置
		config.SaveToFile(""),
	}

	// + 1 保存
	config.G.LaunchesCount++
	if config.G.LaunchesCount == 1 {
		cmds = append(
			cmds,
			messages.Post(messages.RedirectPageRequest{ContentModel: routes.HelpModel}),
			messages.Post(messages.ShowAutoTipsRequest{Text: "首次启动请先仔细阅读帮助页"}),
		)
		return tea.Sequence(cmds...)
	}

	// 没 token 去配置页面
	if config.G.Token == "" {
		cmds = append(
			cmds,
			messages.Post(messages.RedirectPageRequest{ContentModel: routes.SettingModel}),
			messages.Post(messages.ShowAutoTipsRequest{Text: "请先按照说明配置秘钥和节点"}),
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
			messages.Post(messages.LoadingGetToken.Start), api.Client.GetToken,
			messages.Post(messages.LoadingGetToken.End),
		),
	)
	return tea.Sequence(cmds...)
}
