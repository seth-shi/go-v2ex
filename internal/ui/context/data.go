package context

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/resources"
)

type Data struct {
	ScreenHeight  int
	ScreenWidth   int
	ContentHeight int
	// 首页列表数据第几行
	TopicIndex int
	TopicPage  int
	Topics     []*resources.TopicResource

	// 获取配置
	Config config.FileConfig

	// 错误 & 加载
	Error       error
	LoadingText *string

	// 数据相关
	Me *resources.MemberResult
}

func (c *Data) OnWindowChange(msg tea.WindowSizeMsg) {
	c.ScreenWidth = msg.Width
	c.ScreenHeight = msg.Height
}
