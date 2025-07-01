package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/api"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/pkg"
	"github.com/seth-shi/go-v2ex/internal/ui"
)

func init() {
	carbon.SetLayout(carbon.DateTimeLayout)
	carbon.SetTimezone(carbon.PRC)
}
func main() {

	conf, err := config.LoadFileConfig()
	pkg.SetupLogger(&conf)
	api.SetUpHttpClient(&conf)
	pkg.SetUpImageHttpClient(&conf)

	lo.Must1(tea.NewProgram(ui.NewModel(appVersion, err), tea.WithMouseCellMotion()).Run())
}
