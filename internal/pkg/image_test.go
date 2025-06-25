package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImage(t *testing.T) {
	content := "https://i.imgur.com/xKZYQ5j.png\n    https://i.imgur.com/D9KjfGA.jpg"

	links := ExtractImgURLs(content)
	require.Equal(t, 2, len(links))
}
