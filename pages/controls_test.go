package pages

import (
	"strconv"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/stretchr/testify/require"
)

func TestBrowseShortcutsKeepRefreshSeparateFromImageDecode(t *testing.T) {
	refresh := tea.KeyMsg{Type: tea.KeyCtrlR}
	decode := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	history := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}

	require.True(t, key.Matches(refresh, consts.AppKeyMap.Refresh))
	require.False(t, key.Matches(decode, consts.AppKeyMap.Refresh))
	require.True(t, key.Matches(decode, consts.AppKeyMap.KeyR))
	require.True(t, key.Matches(history, consts.AppKeyMap.History))
}

func TestFooterHintsFitAvailableSpaceOnEveryBrowsePage(t *testing.T) {
	for _, page := range []string{"topicPage", "detailPage", "historyPage", "settingPage"} {
		for _, width := range []int{0, 12, 20, 28, 40, 60, 80, 100, 120} {
			t.Run(page+"/"+strconv.Itoa(width), func(t *testing.T) {
				hints := footerHints(page, width)
				require.LessOrEqual(t, lipgloss.Width(hints), width)
				if page == "historyPage" && width >= 20 {
					require.Contains(t, hints, "返回")
				}
			})
		}
	}

	wideTopics := footerHints("topicPage", 120)
	for _, label := range []string{"↑↓", "Enter", "←→", "Tab", "Ctrl+R", "历史", "设置"} {
		require.Contains(t, wideTopics, label)
	}
	wideDetail := footerHints("detailPage", 120)
	for _, label := range []string{"Ctrl+R", "解码", "F1", "历史", "返回"} {
		require.Contains(t, wideDetail, label)
	}
	require.NotContains(t, footerHints("settingPage", 120), "刷新")
	require.NotContains(t, footerHints("settingPage", 120), "历史")
}

func TestHelpFitsWithRefreshAndLocalHistory(t *testing.T) {
	for _, size := range []tea.WindowSizeMsg{{Width: 40, Height: 24}, {Width: 60, Height: 28}, {Width: 100, Height: 32}} {
		t.Run(strconv.Itoa(size.Width), func(t *testing.T) {
			g.Window.SetSize(size)
			view := newHelpPage().View()
			assertViewFits(t, view, size.Width, size.Height)
			require.Contains(t, view, "Ctrl+R")
			require.Contains(t, view, "历史")
		})
	}
}
