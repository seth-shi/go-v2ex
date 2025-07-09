package pkg

import (
	"os"
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/muesli/termenv"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/g"
	"golang.org/x/term"
)

var (
	renderer *glamour.TermRenderer
	once     sync.Once
)

func getRenderer(w int) *glamour.TermRenderer {

	once.Do(
		func() {
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

			// 主要文字颜色
			conf.Document.StylePrimitive.Color = lo.ToPtr("#000000")
			renderer, _ = glamour.NewTermRenderer(
				glamour.WithBaseURL("https://www.v2ex.com"),
				glamour.WithEmoji(),
				glamour.WithWordWrap(w),
				glamour.WithStyles(conf),
			)
		},
	)

	return renderer
}

func SafeRenderHtml(input string) string {

	var (
		w, _ = g.Window.GetSize()
	)

	out, err := getRenderer(w).Render(input)
	if err != nil {
		return input
	}

	return out
}
