package ui

import (
	"reflect"
	"strings"

	"github.com/samber/lo"
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
		return m, m.onConfigLoaded(msgType.Error)
	case messages.RedirectPageRequest:
		m.contentModel = msgType.Page
		// 先切换到列表页面, 再发送消息去请求数据
		var cmd tea.Cmd
		if reflect.DeepEqual(m.contentModel, routes.TopicsModel) {
			cmd = messages.Post(messages.GetTopicsRequest{Page: 1})
		} else if reflect.DeepEqual(m.contentModel, routes.DetailModel) {
			cmd = messages.Post(messages.GetDetailRequest{ID: 1})
		}
		return m, tea.Sequence(messages.Post(messages.ShowTipsRequest{Text: ""}), cmd)
	case tea.KeyMsg:
		switch {
		case key.Matches(msgType, consts.AppKeyMap.SettingPage):
			return m, messages.Post(messages.RedirectPageRequest{Page: lo.If[tea.Model](reflect.DeepEqual(m.contentModel, routes.SettingModel), routes.TopicsModel).Else(routes.SettingModel)})
		case key.Matches(msgType, consts.AppKeyMap.HelpPage):
			return m, messages.Post(messages.RedirectPageRequest{Page: lo.If[tea.Model](reflect.DeepEqual(m.contentModel, routes.HelpModel), routes.TopicsModel).Else(routes.HelpModel)})
		//case key.Matches(msgType, consts.AppKeyMap.Back):
		//	return m, messages.Post(messages.RedirectPageRequest{Page: routes.TopicsModel})
		case key.Matches(msgType, consts.AppKeyMap.SwitchShowMode):
			config.G.SwitchShowMode()
			return m, config.SaveToFile
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

func (m Model) onConfigLoaded(err error) tea.Cmd {

	if err != nil {
		return messages.Post(err)
	}

	// 把配置注入到其他页面
	api.Client.RefreshConfig()
	routes.SettingModel.RefreshConfig()

	return messages.Post(messages.RedirectPageRequest{Page: routes.DetailModel})

	// 第一次没 token 去配置页面
	if config.G.Token == "" {
		return messages.Post(messages.RedirectPageRequest{Page: routes.SettingModel})
	}

	// 去触发对应的地方获取数据
	return tea.Sequence(
		messages.Post(messages.RedirectPageRequest{Page: routes.TopicsModel}),
		// 获取个人信息
		tea.Sequence(messages.Post(messages.LoadingGetToken.Start), api.Client.GetToken, messages.Post(messages.LoadingGetToken.End)),
	)
}
