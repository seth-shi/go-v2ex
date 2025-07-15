package styles

import (
	"strconv"
)

func HotText(val int) string {

	textVal := strconv.Itoa(val)
	switch {
	case val > 100:
		return Err.Render(textVal)
	case val > 10:
		return Active.Render(textVal)
	}

	return Hint.Render(textVal)
}
