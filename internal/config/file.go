package config

import (
	"encoding/json"
	"errors"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mcuadros/go-defaults"
	"github.com/seth-shi/go-v2ex/internal/types"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

func LoadFileConfig() tea.Msg {

	var result = messages.LoadConfigResult{
		Config: types.FileConfig{},
	}
	defer func() {
		defaults.SetDefaults(&result.Config)
	}()

	bf, err := os.ReadFile(result.Config.ConfigPath())
	if err != nil {

		if errors.Is(err, os.ErrNotExist) {
			return result
		}

		result.Error = err
		return result
	}

	result.Error = json.Unmarshal(bf, &result.Config)
	return result
}
