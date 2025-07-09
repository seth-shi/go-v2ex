package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/pages"
)

func init() {
	carbon.SetLayout(carbon.DateTimeLayout)
	carbon.SetTimezone(carbon.PRC)
}
func main() {

	// 初始化到开平页面
	s := pages.NewUI(appVersion)
	lo.Must1(tea.NewProgram(s, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run())
}
