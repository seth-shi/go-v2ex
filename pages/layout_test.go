package pages

import (
	"strconv"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/seth-shi/go-v2ex/v2/consts"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/stretchr/testify/require"
)

func TestTopicPageFitsCommonTerminalWidths(t *testing.T) {
	page := newTopicPage()
	page.loading = false
	page.totalPages = 12
	page.topics = make([]response.TopicResult, 30)
	for i := range page.topics {
		page.topics[i] = response.TopicResult{
			Id:          1,
			Title:       "这是一个很长的主题标题，用来确认在窄终端中也会整齐截断而不会破坏布局 " + strconv.Itoa(i+1),
			Replies:     128,
			Visits:      12_345,
			LastTouched: 1_756_092_800,
			Node:        response.NodeInfoResult{Title: "程序员"},
			Member:      response.MemberResult{Username: "terminal_user"},
		}
	}

	for _, width := range []int{40, 60, 80, 120} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			g.Window.SetSize(tea.WindowSizeMsg{Width: width, Height: 30})
			assertViewFits(t, page.View(), width, 30)
		})
	}
}

func TestTopicTableKeepsImportantMetadata(t *testing.T) {
	page := newTopicPage()
	page.loading = false
	page.topics = []response.TopicResult{{
		Title:       "保留主题元信息",
		Replies:     128,
		Visits:      12_345,
		LastTouched: 1_756_092_800,
		Node:        response.NodeInfoResult{Title: "程序员"},
		Member:      response.MemberResult{Username: "seth"},
	}}

	view := page.renderTopicTable(100, 0)
	require.Contains(t, view, "用户")
	require.Contains(t, view, "seth")
	require.NotContains(t, view, "浏览")

	page.topics[0].Member = response.MemberResult{}
	page.topics[0].LastReplyBy = "reply_user"
	fallbackView := page.renderTopicTable(100, 0)
	require.Contains(t, fallbackView, "↩ reply_user")
}

func TestDetailPageResizesViewport(t *testing.T) {
	page := newDetailPage(1)
	model, _ := page.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	resized := model.(detailPage)
	require.Equal(t, 60, resized.viewport.Width)
	require.GreaterOrEqual(t, resized.viewport.Height, 1)
}

func TestDetailContentFitsResponsiveReadingColumn(t *testing.T) {
	page := newDetailPage(1)
	page.contentDetail = response.V2DetailResult{
		Id:      1,
		Title:   "这是一个用于验证详情页长标题在各种终端宽度下都不会溢出的主题",
		Content: "正文包含一些中文内容和 https://www.v2ex.com 链接。",
		Replies: 2,
		Created: 1_756_092_800,
		Member:  response.MemberResult{Id: 1, Username: "author"},
		Node:    response.NodeInfoResult{Title: "程序员"},
		Supplements: []response.SupplementResult{{
			Content: "这是一条附言，用来验证独立附言卡片。",
			Created: 1_756_093_000,
		}},
	}
	page.contentReply = []response.V2ReplyResult{
		{Content: "第一条评论", Created: 1_756_093_100, Member: response.MemberResult{Id: 1, Username: "author"}},
		{Content: "第二条评论内容稍微长一些", Created: 1_756_093_200, Member: response.MemberResult{Id: 2, Username: "reply_user"}},
	}
	page.replyPageInfo = response.V2PageResponse{CurrPage: 1, TotalPages: 2, TotalCount: 2}

	for _, width := range []int{40, 60, 120} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			g.Window.SetSize(tea.WindowSizeMsg{Width: width, Height: 30})
			page.viewport.Width = width
			assertLinesFit(t, page.headerView(), width)
			assertLinesFit(t, page.buildContent(), width)
			require.Contains(t, page.buildContent(), "Enter  加载下一页评论")
		})
	}
}

func TestDetailReplyStatusCoversEmptyMoreAndFinished(t *testing.T) {
	page := newDetailPage(1)
	page.contentDetail.Replies = 0
	require.Contains(t, page.replyStatus(40), "暂无回复")

	page.contentDetail.Replies = 42
	page.contentReply = make([]response.V2ReplyResult, 20)
	page.replyPageInfo = response.V2PageResponse{CurrPage: 1, TotalPages: 3}
	require.Contains(t, page.replyStatus(40), "Enter  加载下一页评论")

	page.contentReply = make([]response.V2ReplyResult, 42)
	page.replyPageInfo = response.V2PageResponse{CurrPage: 3, TotalPages: 3}
	require.Contains(t, page.replyStatus(40), "已加载全部评论")
}

func TestFooterFitsCommonTerminalWidths(t *testing.T) {
	g.Session.HideFooter.Store(false)
	footer := NewFooter("2.3.9")
	footer.statusBar.FirstColumn = "12/128 • 1280条"
	for _, width := range []int{60, 80, 120} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			g.Window.SetSize(tea.WindowSizeMsg{Width: width, Height: 24})
			assertLinesFit(t, footer.View(), width)
		})
	}
}

func TestTopicFooterGuidesUsersToSettings(t *testing.T) {
	hints := footerHints("topicPage", 100)
	require.Contains(t, hints, "设置")
	require.Contains(t, consts.AppKeyMap.SettingPage.Keys(), ",")
	require.Contains(t, consts.AppKeyMap.SettingPage.Keys(), "`")
}

func TestCompactDetailFooterKeepsReturnHint(t *testing.T) {
	hints := footerHints("detailPage", 40)
	require.Contains(t, hints, "Q")
	require.Contains(t, hints, "返回")
}

func TestEmptyTopicPageHasStateAndCannotOpenZeroID(t *testing.T) {
	g.Window.SetSize(tea.WindowSizeMsg{Width: 60, Height: 24})
	page := newTopicPage()
	page.loading = false
	require.Contains(t, page.View(), "这里还没有主题")

	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Nil(t, cmd)
	require.Equal(t, page, updated)
}

func TestTopicLoadingStateFitsInsidePage(t *testing.T) {
	g.Window.SetSize(tea.WindowSizeMsg{Width: 40, Height: 24})
	page := newTopicPage()
	assertViewFits(t, page.View(), 40, 24)
}

func TestSettingsNodeInputAcceptsCommaSeparator(t *testing.T) {
	page := newSettingPage()
	page.loaded = true
	page.focusIndex = 1
	page.inputs[0].Blur()
	page.inputs[1].Focus()
	page.inputs[1].SetValue("programmer")

	updated, _ := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{','}})
	require.Equal(t, "programmer,", updated.(settingPage).inputs[1].Value())
}

func TestRelativeTimeUsesChineseLocale(t *testing.T) {
	require.Equal(t, "刚刚", relativeTime(time.Now().Unix()))
	result := relativeTime(time.Now().Add(-2 * time.Minute).Unix())
	require.Contains(t, result, "前")
	require.NotContains(t, result, "ago")
}

func TestWelcomeViewFitsCommonTerminalSizes(t *testing.T) {
	for _, size := range []tea.WindowSizeMsg{
		{Width: 40, Height: 24},
		{Width: 60, Height: 28},
		{Width: 100, Height: 32},
	} {
		name := strconv.Itoa(size.Width) + "x" + strconv.Itoa(size.Height)
		t.Run(name, func(t *testing.T) {
			g.Window.SetSize(size)
			view := welcomeView()
			assertViewFits(t, view, size.Width, size.Height)
			require.Contains(t, view, "令牌是可选项")
			require.Contains(t, view, "直接浏览")
		})
	}
}

func TestMainPageDoesNotRequireToken(t *testing.T) {
	require.Len(t, mainPageCommands(&model.FileConfig{}), 1)
	require.Len(t, mainPageCommands(&model.FileConfig{Token: "configured-token"}), 2)
}

func TestTerminalLinkUsesOSC8(t *testing.T) {
	link := terminalLink(tokenSettingsURL, "创建令牌")
	require.Contains(t, link, tokenSettingsURL)
	require.Contains(t, link, "创建令牌")
	require.Equal(t, 8, lipgloss.Width(link))
}

func TestSettingsLinksToNodeDirectory(t *testing.T) {
	g.Window.SetSize(tea.WindowSizeMsg{Width: 80, Height: 24})
	settings := newSettingPage()
	settings.loaded = true
	view := settings.View()
	require.Contains(t, view, nodeDirectoryURL)
	require.Contains(t, view, "查看全部节点")
}

func TestAuxiliaryPagesFitCommonTerminalSizes(t *testing.T) {
	originalJobs := runningJobs
	runningJobs = []string{"github.com/example/a", "github.com/example/a-very-long-dependency-name-that-must-be-truncated"}
	t.Cleanup(func() { runningJobs = originalJobs })

	settings := newSettingPage()
	settings.loaded = true
	settings.inputs[0].SetValue("test-token")
	settings.inputs[1].SetValue("programmer,qna,jobs")

	for _, size := range []tea.WindowSizeMsg{
		{Width: 40, Height: 24},
		{Width: 60, Height: 28},
		{Width: 100, Height: 32},
	} {
		name := strconv.Itoa(size.Width) + "x" + strconv.Itoa(size.Height)
		t.Run(name, func(t *testing.T) {
			g.Window.SetSize(size)
			assertViewFits(t, settings.View(), size.Width, size.Height)
			assertViewFits(t, newHelpPage().View(), size.Width, size.Height)
			assertViewFits(t, newBossPage().View(), size.Width, size.Height)
			assertViewFits(t, loadingView("正在载入配置"), size.Width, size.Height)
		})
	}
}

func assertLinesFit(t *testing.T, view string, width int) {
	t.Helper()
	for _, line := range strings.Split(view, "\n") {
		require.LessOrEqual(t, lipgloss.Width(line), width, "line overflowed: %q", line)
	}
}

func assertViewFits(t *testing.T, view string, width, height int) {
	t.Helper()
	assertLinesFit(t, view, width)
	require.LessOrEqual(t, lipgloss.Height(view), height, "view height overflowed:\n%s", view)
}
