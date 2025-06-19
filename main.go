package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/pkg"
	"github.com/seth-shi/go-v2ex/internal/ui"
)

func init() {
	carbon.SetLayout(carbon.DateTimeLayout)
	carbon.SetTimezone(carbon.PRC)
}
func main() {

	if len(os.Getenv("DEBUG")) > 0 {
		f := lo.Must1(tea.LogToFile("debug.log", "debug"))
		defer f.Close()
	} else {
		log.SetOutput(pkg.DiscardLogger().Writer())
	}

	lo.Must1(tea.NewProgram(ui.NewModel(appVersion), tea.WithMouseCellMotion()).Run())
}
