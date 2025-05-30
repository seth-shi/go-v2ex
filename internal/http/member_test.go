package http

import (
	"testing"

	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/stretchr/testify/require"
)

func TestMember(t *testing.T) {
	cfg := config.FileConfig{Token: "1"}
	client := V2exClient.SetConfig(cfg)
	res, err := client.GetMember(t.Context())
	require.ErrorContains(t, err, "Invalid token")
	require.Nil(t, res)
}
