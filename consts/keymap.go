package consts

import (
	"github.com/charmbracelet/bubbles/key"
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
			k.KeyE, k.KeyQ, k.KeyR,
		}, // first column
		{
			k.CtrlQuit, k.HelpPage, k.SettingPage,
			k.UpgradeApp,
			k.SwitchShowMode, k.Space, k.F1,
		}, // second column
	}
}

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
		key.WithHelp("w/↑", "[主题页]列表上一个"),
	),
	Down: key.NewBinding(
		key.WithKeys("s", "down", "j"),
		key.WithHelp("s/↓", "[主题页]列表下一个"),
	),
	Left: key.NewBinding(
		key.WithKeys("a", "left", "h"),
		key.WithHelp("a/←", "[主题页]上一页"),
	),
	Right: key.NewBinding(
		key.WithKeys("d", "right", "l"),
		key.WithHelp("d/→", "[主题页]下一页"),
	),
	KeyQ: key.NewBinding(
		key.WithKeys("q", "H"),
		key.WithHelp("q", "返回上一页"),
	),
	HelpPage: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "查看帮助页面(再按一次返回首页)"),
	),
	SettingPage: key.NewBinding(
		key.WithKeys("`"),
		key.WithHelp("`", "[反引号]进入配置页面(再按一次返回首页)"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "[主题页]下一个节点"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("空格", "老板键"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "[主题页]上一个切点"),
	),
	CtrlQuit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "退出程序"),
	),
	KeyE: key.NewBinding(
		key.WithKeys("e", "enter", "o"),
		key.WithHelp("e/enter", "[主题页]查看主题详情 / [详情页]加载评论"),
	),
	SwitchShowMode: key.NewBinding(
		key.WithKeys("="),
		key.WithHelp("=", "[等于号]切换底部显示隐藏"),
	),
	KeyR: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "[主题页]切换接口版本 / [详情页]加载图片"),
	),
	UpgradeApp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "更新应用(需要网络可以访问 github)"),
	),
	F1: key.NewBinding(
		key.WithKeys("f1"),
		key.WithHelp("f1", "[详情页]打开链接 / [配置页]打开配置文件"),
	),
}
