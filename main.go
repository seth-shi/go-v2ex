package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/pkg"
	"github.com/seth-shi/go-v2ex/internal/ui"
)

func init() {
	carbon.SetLayout(carbon.DateTimeLayout)
	carbon.SetTimezone(carbon.PRC)
}
func main() {

	err := config.LoadFileConfig()
	pkg.SetupLogger(config.G.Debug)

	lo.Must1(tea.NewProgram(ui.NewModel(appVersion, err), tea.WithMouseCellMotion()).Run())
}
