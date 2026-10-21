package pages

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/model"
	"github.com/seth-shi/go-v2ex/v2/response"
	"github.com/seth-shi/go-v2ex/v2/styles"
	"github.com/stretchr/testify/require"
)

var compactTopicTitle = regexp.MustCompile(`topic[0-9]{3}`)
var compactHistoryTitle = regexp.MustCompile(`history[0-9]{3}`)
var compactSGR = regexp.MustCompile("\x1b\\[([0-9;:]*)m")

func compactTopicFixture(count int) topicPage {
	page := newTopicPage()
	page.loading = false
	page.topics = make([]response.TopicResult, count)
	for i := range page.topics {
		page.topics[i] = response.TopicResult{
			Id:          int64(i + 1),
			Title:       fmt.Sprintf("topic%03d 紧凑列表验收", i+1),
			Replies:     10 + i,
			LastTouched: 1_756_092_800,
			Node:        response.NodeInfoResult{Title: "程序员"},
			Member:      response.MemberResult{Username: fmt.Sprintf("u%03d", i+1)},
		}
	}
	return page
}

func compactHistoryFixture(count int) historyPage {
	page := historyPage{entries: make([]model.HistoryEntry, count), loaded: true}
	for i := range page.entries {
		page.entries[i] = model.HistoryEntry{
			TopicID:   int64(i + 1),
			Title:     fmt.Sprintf("history%03d 紧凑历史验收", i+1),
			Node:      "程序员",
			Author:    fmt.Sprintf("u%03d", i+1),
			VisitedAt: 1_756_092_800,
		}
	}
	return page
}

func compactDataLines(view string, titlePattern *regexp.Regexp) []string {
	var rows []string
	for _, line := range strings.Split(ansi.Strip(view), "\n") {
		if titlePattern.MatchString(line) {
			rows = append(rows, line)
		}
	}
	return rows
}

func TestCompactTopicsUseOneTableRowAtDesktopWidths(t *testing.T) {
	useTemporaryHistory(t)
	for _, width := range []int{80, 100, 120} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			g.Window.SetSize(tea.WindowSizeMsg{Width: width, Height: 24})
			page := compactTopicFixture(30)
			view := ansi.Strip(page.View())
			assertViewFits(t, view, width, 23)
			for _, column := range []string{"节点", "标题", "作者", "回复", "活跃"} {
				require.Contains(t, view, column)
			}
			rows := compactDataLines(view, compactTopicTitle)
			require.GreaterOrEqual(t, len(rows), 15, "a 24-line desktop terminal should display dense single-line rows")
			for _, row := range rows {
				title := compactTopicTitle.FindString(row)
				index, err := strconv.Atoi(strings.TrimPrefix(title, "topic"))
				require.NoError(t, err)
				require.Contains(t, row, "程序员", "node and title must share one table row")
				require.Contains(t, row, fmt.Sprintf("u%03d", index), "author must be on the title's row")
				require.Contains(t, row, strconv.Itoa(9+index), "reply count must be on the title's row")
			}
		})
	}
}

func TestCompactTopicsUseAdjacentTwoLineRowsOnNarrowTerminals(t *testing.T) {
	useTemporaryHistory(t)
	for _, width := range []int{40, 60} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			g.Window.SetSize(tea.WindowSizeMsg{Width: width, Height: 24})
			view := ansi.Strip(compactTopicFixture(30).renderTopics(18))
			assertDenseTwoLineRows(t, view, compactTopicTitle)
			assertViewFits(t, view, width, 18)
		})
	}
}

func assertDenseTwoLineRows(t *testing.T, view string, titlePattern *regexp.Regexp) {
	t.Helper()
	lines := strings.Split(view, "\n")
	first, last, count := -1, -1, 0
	for i, line := range lines {
		if titlePattern.MatchString(line) {
			if first < 0 {
				first = i
			}
			last = i
			count++
		}
	}
	require.Greater(t, count, 1)
	require.Less(t, last+1, len(lines), "the final title needs a complete metadata line")
	require.Equal(t, count*2, last-first+2, "narrow rows must occupy exactly two lines without blank separators")
	for _, line := range lines[first : last+2] {
		require.NotEmpty(t, strings.TrimSpace(line), "blank separators reduce the requested compact density")
	}
	for i := first; i <= last; i += 2 {
		require.True(t, titlePattern.MatchString(lines[i]))
		require.Contains(t, lines[i+1], "程序员")
		require.Contains(t, lines[i+1], "u")
	}
}

func TestCompactTopicFirstAndLastSelectionRemainVisible(t *testing.T) {
	useTemporaryHistory(t)
	for _, width := range []int{40, 80} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			g.Window.SetSize(tea.WindowSizeMsg{Width: width, Height: 24})
			page := compactTopicFixture(30)
			view := ansi.Strip(page.View())
			assertViewFits(t, view, width, 23)
			require.Contains(t, view, "topic001")
			require.Contains(t, view, "u001")
			updated, command := page.Update(tea.KeyMsg{Type: tea.KeyUp})
			require.Nil(t, command)
			page = updated.(topicPage)
			require.Equal(t, 29, page.index)
			view = ansi.Strip(page.View())
			assertViewFits(t, view, width, 23)
			require.Contains(t, view, "topic030")
			require.Contains(t, view, "u030", "the last selected item's metadata must remain visible")
			require.NotContains(t, view, "topic001")
		})
	}
}

func TestCompactHistoryUsesTableAndDenseNarrowRows(t *testing.T) {
	useTemporaryHistory(t)
	for _, width := range []int{40, 60, 80, 100} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			g.Window.SetSize(tea.WindowSizeMsg{Width: width, Height: 24})
			page := compactHistoryFixture(30)
			view := ansi.Strip(page.View())
			assertViewFits(t, view, width, 23)
			if width < 80 {
				assertDenseTwoLineRows(t, view, compactHistoryTitle)
			} else {
				for _, column := range []string{"节点", "标题", "作者", "最近打开"} {
					require.Contains(t, view, column)
				}
				rows := compactDataLines(view, compactHistoryTitle)
				require.GreaterOrEqual(t, len(rows), 15)
				for _, row := range rows {
					title := compactHistoryTitle.FindString(row)
					index, err := strconv.Atoi(strings.TrimPrefix(title, "history"))
					require.NoError(t, err)
					require.Contains(t, row, "程序员")
					require.Contains(t, row, fmt.Sprintf("u%03d", index))
				}
			}
			updated, _ := page.Update(tea.KeyMsg{Type: tea.KeyUp})
			page = updated.(historyPage)
			view = ansi.Strip(page.View())
			assertViewFits(t, view, width, 23)
			require.Contains(t, view, "history030")
			require.Contains(t, view, "u030")
		})
	}
}

func TestCompactSelectedRowsUseFullHighlightIncludingReadTopics(t *testing.T) {
	useTemporaryHistory(t)
	originalProfile, originalDark := lipgloss.ColorProfile(), lipgloss.HasDarkBackground()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(originalProfile)
		lipgloss.SetHasDarkBackground(originalDark)
	})
	for _, dark := range []bool{false, true} {
		for _, width := range []int{40, 80} {
			for _, read := range []bool{false, true} {
				name := fmt.Sprintf("dark=%t/width=%d/read=%t", dark, width, read)
				t.Run(name, func(t *testing.T) {
					lipgloss.SetHasDarkBackground(dark)
					g.Window.SetSize(tea.WindowSizeMsg{Width: width, Height: 24})
					g.History = model.NewHistoryStore(filepath.Join(t.TempDir(), "history.json"))
					if read {
						require.NoError(t, g.History.Record(model.HistoryEntry{TopicID: 2, Title: "read"}))
					}
					page := compactTopicFixture(3)
					page.index = 1
					assertSelectedHighlightRows(t, page.View(), "topic002", width, width < 80, dark)
					if !read {
						history := compactHistoryFixture(3)
						history.index = 1
						assertSelectedHighlightRows(t, history.View(), "history002", width, width < 80, dark)
					}
				})
			}
		}
	}
}

func assertSelectedHighlightRows(t *testing.T, view, title string, width int, twoLines, dark bool) {
	t.Helper()
	color := styles.Theme.Selected.Light
	if dark {
		color = styles.Theme.Selected.Dark
	}
	rgb, err := strconv.ParseUint(strings.TrimPrefix(color, "#"), 16, 24)
	require.NoError(t, err)
	expected := fmt.Sprintf("rgb:%d:%d:%d", rgb>>16, (rgb>>8)&255, rgb&255)
	lines := strings.Split(view, "\n")
	selected := -1
	for i, line := range lines {
		if strings.Contains(ansi.Strip(line), title) {
			selected = i
			break
		}
	}
	require.GreaterOrEqual(t, selected, 0)
	rowCount := 1
	if twoLines {
		rowCount = 2
	}
	contentWidth, _ := listGeometry(width)
	for _, line := range lines[selected : selected+rowCount] {
		colored, missing := compactBackgroundCoverage(line, expected)
		require.Empty(t, missing, "selected text lost its highlight background: %q", line)
		require.GreaterOrEqual(t, colored, contentWidth, "the entire selected content row, including padding, needs highlight background")
	}
}

// Track rendered SGR state rather than merely searching for one background
// escape: nested foreground styles can reset the selection mid-row.
func compactBackgroundCoverage(line, expected string) (colored int, missing []string) {
	background := ""
	position := 0
	consume := func(text string) {
		text = ansi.Strip(text)
		if background == expected {
			colored += lipgloss.Width(text)
		} else if strings.TrimSpace(text) != "" {
			missing = append(missing, text)
		}
	}
	for _, match := range compactSGR.FindAllStringSubmatchIndex(line, -1) {
		consume(line[position:match[0]])
		parameters := strings.Split(strings.ReplaceAll(line[match[2]:match[3]], ":", ";"), ";")
		for i := 0; i < len(parameters); i++ {
			code, _ := strconv.Atoi(parameters[i])
			switch {
			case code == 0 || code == 49:
				background = ""
			case code == 38 || code == 48 || code == 58:
				if i+4 < len(parameters) && parameters[i+1] == "2" {
					if code == 48 {
						background = "rgb:" + strings.Join(parameters[i+2:i+5], ":")
					}
					i += 4
				} else if i+2 < len(parameters) && parameters[i+1] == "5" {
					if code == 48 {
						background = "index:" + parameters[i+2]
					}
					i += 2
				}
			case code >= 40 && code <= 47 || code >= 100 && code <= 107:
				background = fmt.Sprintf("ansi:%d", code)
			}
		}
		position = match[1]
	}
	consume(line[position:])
	return colored, missing
}
