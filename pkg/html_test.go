package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRendererCacheRespectsRequestedWidth(t *testing.T) {
	narrow := getRenderer(40)
	wide := getRenderer(80)
	require.NotSame(t, narrow, wide)
	require.Same(t, narrow, getRenderer(40))
}
