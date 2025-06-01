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
	"github.com/seth-shi/go-v2ex/internal/ui/components/header"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
	"github.com/seth-shi/go-v2ex/internal/ui/routes"
)

type Model struct {
	currBodyModel tea.Model
	headerModel   tea.Model
	footerModel   tea.Model
}

func NewModel() Model {
	return Model{
		currBodyModel: routes.SplashModel,
		headerModel:   header.New(),
		footerModel:   footer.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		// 加载配置
		config.LoadFileConfig,
		// 其它不要用 init 初始化, 使用消息去刷新
		m.headerModel.Init(),
		m.currBodyModel.Init(),
		m.footerModel.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msgType := msg.(type) {
	// 全局监听, 让子组件更新
	case tea.WindowSizeMsg:
		config.Screen.Width = msgType.Width
		config.Screen.Height = msgType.Height
	// 全局监听, 无需转给子组件
	case messages.LoadConfigResult:
		return m, m.onConfigLoaded(msgType.Error)
	case messages.SettingSaveResult:
		return m, m.onConfigLoaded(nil)
	case messages.RedirectPageRequest:
		m.currBodyModel = msgType.Page
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msgType, consts.AppKeyMap.SettingPage):
			return m, messages.Post(messages.RedirectPageRequest{Page: lo.If[tea.Model](reflect.DeepEqual(m.currBodyModel, routes.SettingModel), routes.TopicsModel).Else(routes.SettingModel)})
		case key.Matches(msgType, consts.AppKeyMap.HelpPage):
			return m, messages.Post(messages.RedirectPageRequest{Page: lo.If[tea.Model](reflect.DeepEqual(m.currBodyModel, routes.HelpModel), routes.TopicsModel).Else(routes.HelpModel)})
		case key.Matches(msgType, consts.AppKeyMap.Back):
			return m, messages.Post(messages.RedirectPageRequest{Page: routes.TopicsModel})
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
	m.headerModel, cmd = m.headerModel.Update(msg)
	cmds = append(cmds, cmd)
	m.currBodyModel, cmd = m.currBodyModel.Update(msg)
	cmds = append(cmds, cmd)
	m.footerModel, cmd = m.footerModel.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	var (
		output strings.Builder
	)

	output.WriteString(m.headerModel.View())
	output.WriteRune('\n')
	output.WriteString(m.currBodyModel.View())
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

	// 第一次没 token 去配置页面
	if config.G.Token == "" {
		return messages.Post(messages.RedirectPageRequest{Page: routes.SettingModel})
	}

	// 去触发对应的地方获取数据
	return tea.Batch(
		messages.Post(messages.RedirectPageRequest{Page: routes.TopicsModel}),
		messages.Post(messages.GetTopicsRequest{Page: 1}),
		messages.Post(messages.GetMeRequest{}),
	)
}
