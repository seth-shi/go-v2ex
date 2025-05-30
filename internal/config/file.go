package config

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/go-homedir"
)

type FileConfig struct {
	Token   string `json:"personal_access_token"`
	Nodes   string `json:"nodes" default:"latest,hot"`
	Timeout uint   `json:"timeout" default:"5"`
}

func NewConfig() (*FileConfig, error) {

	var cfg = FileConfig{}
	defer func() {
		defaults.SetDefaults(&cfg)
	}()

	bf, err := os.ReadFile(cfg.ConfigPath())
	if err != nil {

		if errors.Is(err, os.ErrNotExist) {
			return &cfg, nil
		}

		return &cfg, err
	}

	if err := json.Unmarshal(bf, &cfg); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}

func (c *FileConfig) SaveToFile() error {

	bytesData, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(c.ConfigPath(), bytesData, 0644)
}

func (c *FileConfig) ConfigPath() string {
	home, err := homedir.Dir()
	if err != nil {
		home = "."
	}

	return path.Join(home, ".go_v2ex.json")
}
