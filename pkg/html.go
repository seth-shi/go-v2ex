package pkg

import (
	"os"
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

var (
	renderer *glamour.TermRenderer
	once     sync.Once
)

func getRenderer(w int) *glamour.TermRenderer {

	once.Do(
		func() {
			// 主要文字颜色
			renderer, _ = glamour.NewTermRenderer(
				glamour.WithBaseURL("https://www.v2ex.com"),
				glamour.WithEmoji(),
				glamour.WithWordWrap(w),
				glamour.WithStyles(getDefaultStyle()),
			)
		},
	)

	return renderer
}

func getDefaultStyle() ansi.StyleConfig {
	var conf = ansi.StyleConfig{}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		conf = styles.NoTTYStyleConfig
	} else {
		if termenv.HasDarkBackground() {
			conf = styles.DarkStyleConfig
		} else {
			conf = styles.LightStyleConfig
		}
	}

	// 设置为 nil, 不要用设置文字颜色
	conf.Document.Color = nil

	return conf
}

func SafeRenderHtml(input string, w int) string {

	out, err := getRenderer(w).Render(input)
	if err != nil {
		return input
	}

	return out
}
