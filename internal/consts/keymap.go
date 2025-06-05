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
	Quit           key.Binding
	Tab            key.Binding
	Back           key.Binding
	ShiftTab       key.Binding
	Enter          key.Binding
	SwitchShowMode key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Left, k.Right, k.Tab, k.ShiftTab, k.Enter, // first column
		k.HelpPage, k.SettingPage, k.Quit, // second column
		k.SwitchShowMode,
	}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Tab, k.ShiftTab, k.Enter, k.Back}, // first column
		{k.Quit, k.HelpPage, k.SettingPage, k.SwitchShowMode},               // second column
	}
}

var AppKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("w", "up"),
		key.WithHelp("w/↑", "[主题页]移动到上一个"),
	),
	Down: key.NewBinding(
		key.WithKeys("s", "down"),
		key.WithHelp("s/↓", "[主题页]列表移动到下一个"),
	),
	Left: key.NewBinding(
		key.WithKeys("a", "left"),
		key.WithHelp("a/←", "[主题页]上一页"),
	),
	Right: key.NewBinding(
		key.WithKeys("d", "right"),
		key.WithHelp("d/→", "[主题页]下一页"),
	),
	Back: key.NewBinding(
		key.WithKeys("q"),
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
		key.WithHelp("tab", "[主题页]切换下一个节点"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "[主题页]切换上一个切点"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "退出程序"),
	),
	Enter: key.NewBinding(
		key.WithKeys("e", "enter"),
		key.WithHelp("e/enter", "[主题页]查看主题详情"),
	),
	SwitchShowMode: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "[减号]切换底部显示隐藏"),
	),
}
