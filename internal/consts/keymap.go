package consts

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Help    key.Binding
	Setting key.Binding
	Quit    key.Binding
	Tab     key.Binding
	Back    key.Binding
	Enter   key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right}, // first column
		{k.Help, k.Setting, k.Quit},     // second column
		{k.Tab, k.Enter},                // second column
	}
}

var AppKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "w"),
		key.WithHelp("↑ / w", "列表移动到上一个"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "s"),
		key.WithHelp("↓ / s", "列表移动到下一个"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "a"),
		key.WithHelp("← / a", "上一页"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "d"),
		key.WithHelp("→ / d", "下一页"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp(" ?", "查看帮助页面"),
	),
	Back: key.NewBinding(
		key.WithKeys("ctrl+b", "backspace"),
		key.WithHelp("ctrl+b / 删除键", "返回上一页"),
	),
	Setting: key.NewBinding(
		key.WithKeys("`"),
		key.WithHelp("`", "反引号进入配置页面"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "切换主题"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc / ctrl+c", "退出程序"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "查看主题详情"),
	),
}
