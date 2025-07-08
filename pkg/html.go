package pkg

import (
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
)

func SafeRenderHtml(input string) string {

	markdown, err := htmltomarkdown.ConvertString(
		input,
		converter.WithDomain("https://www.v2ex.com"),
	)
	if err != nil {
		return input
	}

	out, err := glamour.Render(markdown, styles.AsciiStyle)
	if err != nil {
		return input
	}

	return out
}
