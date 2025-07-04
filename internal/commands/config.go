package commands

import (
	"encoding/json"
	"errors"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/internal/config"
	"github.com/seth-shi/go-v2ex/internal/model/messages"
)

func LoadConfig() tea.Cmd {
	return func() tea.Msg {

		cfg := config.NewFileConfig()
		bf, err := os.ReadFile(config.SavePath())
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		err = json.Unmarshal(bf, &cfg)
		if err != nil {
			return err
		}

		return messages.LoadConfigResult{Result: cfg}
	}
}

func SaveToFile(conf *config.FileConfig) error {
	bytesData, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	if err = os.WriteFile(config.SavePath(), bytesData, 0644); err != nil {
		return err
	}

	return nil
}
