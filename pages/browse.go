package pages

import (
	"fmt"
	"strings"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/seth-shi/go-v2ex/v2/g"
	"github.com/seth-shi/go-v2ex/v2/styles"
)

var requestSequence atomic.Uint64

func nextRequestID() uint64 { return requestSequence.Add(1) }
func pageSize() (int, int) {
	w, h := g.Window.GetSize()
	if !g.Session.HideFooter.Load() && g.Config.Get().ShowFooter() {
		h--
	}
	return max(w, 1), max(h, 1)
}
func fitPage(view string, w, h int) string {
	return lipgloss.NewStyle().MaxWidth(max(w, 1)).MaxHeight(max(h, 1)).Render(view)
}
func listGeometry(w int) (int, int) {
	width := max(w-2, 1)
	return width, max((w-width)/2, 0)
}
func cleanLine(value string) string { return strings.Join(strings.Fields(ansi.Strip(value)), " ") }
func joinEdges(left, right string, width int) string {
	right = ansi.Truncate(right, max(width/2, 1), "…")
	left = ansi.Truncate(left, max(width-lipgloss.Width(right)-1, 1), "…")
	return left + strings.Repeat(" ", max(width-lipgloss.Width(left)-lipgloss.Width(right), 0)) + right
}
func indentLines(content string, pad int) string {
	return lipgloss.NewStyle().PaddingLeft(max(pad, 0)).Render(content)
}
func visibleRange(index, count, visible int) (int, int) {
	start := min(max(index-visible/2, 0), max(count-visible, 0))
	return start, min(start+visible, count)
}
func renderListItem(title, meta, count string, width int, selected, read bool) string {
	marker := "  "
	titleStyle := styles.Title
	if read {
		titleStyle = styles.Meta
		marker = "· "
	}
	if selected {
		marker = "▸ "
	}
	contentWidth := max(width-2, 1)
	title = ansi.TruncateWc(title, contentWidth, "…")
	if width >= 54 && count != "" {
		title = joinEdges(title, count, contentWidth)
	} else {
		if count != "" {
			meta = fmt.Sprintf("%s 回复 · %s", count, meta)
		}
	}
	meta = ansi.TruncateWc(meta, contentWidth, "…")
	if selected {
		// Apply selection once to the whole row so nested metadata styles
		// cannot replace the high-contrast foreground or background.
		return styles.SelectedRow.Width(width).Render(marker + title + "\n  " + meta)
	}
	return marker + titleStyle.Render(title) + "\n  " + styles.Meta.Render(meta)
}

const compactListMinWidth = 80

func topicColumnWidths(width int) []int   { return []int{8, max(width-43, 1), 12, 5, 8} }
func historyColumnWidths(width int) []int { return []int{8, max(width-38, 1), 12, 10} }

func compactCell(value string, width int, right bool) string {
	value = ansi.TruncateWc(cleanLine(value), width, "…")
	padding := strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
	if right {
		return padding + value
	}
	return value + padding
}

func compactHeader(labels []string, widths []int, width int) string {
	cells := make([]string, len(labels))
	for i, label := range labels {
		cells[i] = compactCell(label, widths[i], i >= 3)
	}
	return styles.Meta.Bold(true).Background(styles.Theme.Surface).Width(width).Render("  "+strings.Join(cells, "  ")) + "\n" + styles.Divider.Render(strings.Repeat("─", width))
}

func compactRow(values []string, widths []int, width int, selected, read bool) string {
	marker := "  "
	if read {
		marker = "· "
	}
	if selected {
		marker = "▸ "
	}
	cells := make([]string, len(values))
	for i, value := range values {
		cell := compactCell(value, widths[i], i >= 3)
		if !selected {
			style := styles.Meta
			if i == 0 {
				style = styles.Active
			}
			if i == 1 && !read {
				style = lipgloss.NewStyle().Foreground(styles.Theme.Text)
			}
			cell = style.Render(cell)
		}
		cells[i] = cell
	}
	row := marker + strings.Join(cells, "  ")
	if selected {
		return styles.SelectedRow.Width(width).Render(row)
	}
	return row
}
func emptyState(title, hint string, w, h int) string {
	width := max(w-4, 1)
	content := styles.Title.Render(ansi.Hardwrap(title, width, true)) + "\n" + styles.Meta.Render(ansi.Hardwrap(hint, width, true))
	return fitPage(lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content), w, h)
}
func requestErrorView(title string, err error, w, h int) string {
	text := "请求失败，请检查网络或终端代理"
	if err != nil {
		text = cleanLine(err.Error())
	}
	width := min(max(w-6, 1), 76)
	content := styles.Err.Bold(true).Render(ansi.Hardwrap(title, width, true)) + "\n\n" + styles.Meta.Render(ansi.Hardwrap(ansi.TruncateWc(text, width*2, "…"), width, true)) + "\n\n" + styles.Active.Render(ansi.Hardwrap("Ctrl+R 重试 · Q 返回", width, true))
	return fitPage(lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content), w, h)
}

type historyUpdated struct{ Err error }

func loadHistory() tea.Cmd { return func() tea.Msg { return historyUpdated{Err: g.History.Load()} } }
