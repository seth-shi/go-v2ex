package model

import (
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxWidth  = 10000 // 假设最大宽度为10000
	maxHeight = 10000 // 假设最大高度为10000
)

type WindowSize struct {
	value atomic.Int64 // 存储编码后的数值
}

// SetSize 原子性地设置窗口大小
func (w *WindowSize) SetSize(msg tea.WindowSizeMsg) {
	// 确保数值在允许范围内
	width := clamp(msg.Width, 0, maxWidth)
	height := clamp(msg.Height, 0, maxHeight)

	// 编码：value = width + height * maxWidth
	value := int64(width) + int64(height)*maxWidth
	w.value.Store(value)
}

// GetSize 原子性地获取窗口大小
func (w *WindowSize) GetSize() (int, int) {
	value := w.value.Load()
	width := int(value % maxWidth)
	height := int(value / maxWidth)
	return width, height
}

// clamp 将数值限制在 [min, max] 范围内
func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
