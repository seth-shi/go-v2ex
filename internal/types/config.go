package types

import (
	"encoding/json"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
)

type FileConfig struct {
	Token   string `json:"personal_access_token"`
	Nodes   string `json:"nodes" default:"latest,hot"`
	Timeout uint   `json:"timeout" default:"5"`
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

	return path.Join(home, ".go-v2ex.json")
}
