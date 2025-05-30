package ui

import (
	"strings"

	"github.com/seth-shi/go-v2ex/internal/http"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/consts"
	"github.com/seth-shi/go-v2ex/internal/ui/context"
	"github.com/seth-shi/go-v2ex/internal/ui/events"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
	"github.com/seth-shi/go-v2ex/internal/ui/pages/help"
	"github.com/seth-shi/go-v2ex/internal/ui/pages/home"
	"github.com/seth-shi/go-v2ex/internal/ui/pages/setting"
)

type Model struct {
	spinner       spinner.Model
	ctx           *context.Data
	currBodyModel tea.Model
	helpModel     help.Model
	settingModel  setting.Model
	homeModel     home.Model
}

func NewModel() Model {

	ctxData := &context.Data{LoadingText: lo.ToPtr("初始化配置中...")}
	return Model{
		ctx:          ctxData,
		spinner:      spinner.New(spinner.WithSpinner(spinner.Globe)),
		helpModel:    help.New(ctxData),
		settingModel: setting.New(ctxData),
		homeModel:    home.New(ctxData),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.spinner.Tick,
		events.InitFileConfig,
		m.settingModel.Init(),
		m.homeModel.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch typeMsg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ctx.OnWindowChange(typeMsg)
		return m, nil
	case messages.UiMessageInit:
		return m, m.initSuccess(typeMsg)
	case messages.GoToHome:
		m.currBodyModel = m.homeModel
		m.refreshConfig(typeMsg.Config)
		return m, nil
	case messages.GetMe:
		m.ctx.Error = typeMsg.Error
		m.ctx.Me = typeMsg.Member
		m.ctx.LoadingText = nil
		return m, nil
	case messages.GetTopics:
		m.ctx.Error = typeMsg.Error
		m.ctx.Topics = typeMsg.Topics
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(typeMsg)
		return m, cmd
	case tea.KeyMsg:
		switch {
		case key.Matches(typeMsg, consts.AppKeyMap.Setting):
			m.currBodyModel = m.settingModel
			return m, nil
		case key.Matches(typeMsg, consts.AppKeyMap.Help):
			m.currBodyModel = m.helpModel
			return m, nil
		case key.Matches(typeMsg, consts.AppKeyMap.Quit):
			return m, tea.Quit
		}
	}

	return m.bodyUpdate(msg)
}

func (m Model) View() string {

	s := strings.Builder{}
	s.WriteString(m.headerView())
	s.WriteString(m.bodyView())
	s.WriteString(m.footerView())
	return s.String()
}

func (m *Model) initSuccess(typeMsg messages.UiMessageInit) tea.Cmd {

	m.refreshConfig(typeMsg.Config)
	m.ctx.Error = typeMsg.Error
	m.ctx.LoadingText = nil

	if m.ctx.Config.Token == "" {
		m.currBodyModel = m.settingModel
		return nil
	}

	// 否则跳转到首页
	http.V2exClient.SetConfig(m.ctx.Config)
	m.currBodyModel = m.homeModel

	// 获取个人中心的数据
	m.ctx.LoadingText = lo.ToPtr("登录中...")
	m.ctx.TopicPage = 1
	return tea.Batch(events.GetMe, events.GetTopics(1))
}

func (m *Model) refreshConfig(config *config.FileConfig) {

	if config == nil {
		return
	}

	m.ctx.Config = lo.FromPtr(config)
	// 没配置到秘钥跳转到设置也
	m.settingModel.UpdateInputValues()
	// 否则跳转到首页
	http.V2exClient.SetConfig(m.ctx.Config)
}
