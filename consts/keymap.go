package consts

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

type KeyMap struct {
	Up             key.Binding
	Down           key.Binding
	Left           key.Binding
	Right          key.Binding
	HelpPage       key.Binding
	SettingPage    key.Binding
	Space          key.Binding
	CtrlQuit       key.Binding
	Tab            key.Binding
	KeyQ           key.Binding
	ShiftTab       key.Binding
	KeyE           key.Binding
	SwitchShowMode key.Binding
	KeyR           key.Binding
	UpgradeApp     key.Binding
	F1             key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up, k.Down, k.Left, k.Right,
			k.Tab, k.ShiftTab,
			k.KeyE, k.KeyR, k.F1,
		}, // first column
		{
			k.KeyQ, k.CtrlQuit, k.HelpPage, k.SettingPage,
			k.UpgradeApp,
			k.SwitchShowMode, k.Space,
		}, // second column
	}
}

var (
	topicPageTitle   = styles.Active.Bold(true).Render("[主题页]")
	allPageTitle     = styles.Bold.Underline(true).Render("[任意页]")
	detailPageTitle  = styles.Err.Bold(true).Render("[详情页]")
	settingPageTitle = styles.Bold.Underline(true).Render("[配置页]")
)

// AppKeyMap
// vim key bind style : https://github.com/philc/vimium#keyboard-bindings
// ?       show the help dialog for a list of all available keys
// h       scroll left
// j       scroll down
// k       scroll up
// l       scroll right
var AppKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("w", "up", "k"),
		key.WithHelp("w/↑", fmt.Sprintf("%s列表上一个", topicPageTitle)),
	),
	Down: key.NewBinding(
		key.WithKeys("s", "down", "j"),
		key.WithHelp("s/↓", fmt.Sprintf("%s列表下一个", topicPageTitle)),
	),
	Left: key.NewBinding(
		key.WithKeys("a", "left", "h"),
		key.WithHelp("a/←", fmt.Sprintf("%s上一页", topicPageTitle)),
	),
	Right: key.NewBinding(
		key.WithKeys("d", "right", "l"),
		key.WithHelp("d/→", fmt.Sprintf("%s下一页", topicPageTitle)),
	),
	KeyQ: key.NewBinding(
		key.WithKeys("q", "H"),
		key.WithHelp("q", fmt.Sprintf("%s返回上一页", allPageTitle)),
	),
	HelpPage: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", fmt.Sprintf("%s查看帮助页面(再按一次返回首页)", allPageTitle)),
	),
	SettingPage: key.NewBinding(
		key.WithKeys("`"),
		key.WithHelp("`", fmt.Sprintf("%s反引号:查看帮助页面(再按一次返回首页)", allPageTitle)),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", fmt.Sprintf("%s下一个节点", topicPageTitle)),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", fmt.Sprintf("%s上一个切点", topicPageTitle)),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("空格", fmt.Sprintf("%s老板键", allPageTitle)),
	),
	CtrlQuit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", fmt.Sprintf("%s退出程序", allPageTitle)),
	),
	SwitchShowMode: key.NewBinding(
		key.WithKeys("="),
		key.WithHelp("=", fmt.Sprintf("%s等于号:切换底部显示隐藏", allPageTitle)),
	),
	KeyE: key.NewBinding(
		key.WithKeys("e", "enter", "o"),
		key.WithHelp("e/enter", fmt.Sprintf("%s查看主题详情 / %s加载评论", topicPageTitle, detailPageTitle)),
	),
	KeyR: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", fmt.Sprintf("%s切换接口版本 / %s解码内容(图片/base64)", topicPageTitle, detailPageTitle)),
	),
	UpgradeApp: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", fmt.Sprintf("%s更新应用", allPageTitle)),
	),
	F1: key.NewBinding(
		key.WithKeys("f1"),
		key.WithHelp("f1", fmt.Sprintf("%s浏览器打开 / %s打开配置文件", topicPageTitle, settingPageTitle)),
	),
}
