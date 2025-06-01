package config

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/go-homedir"
	"github.com/seth-shi/go-v2ex/internal/ui/messages"
)

const (
	showHeader = 1
	showFooter = 2
	showAll    = 3
	showEmpty  = 4
)

var (
	G = newFileConfig()
)

type fileConfig struct {
	Token      string `json:"personal_access_token"`
	Nodes      string `json:"nodes" default:"latest,hot"`
	Timeout    uint   `json:"timeout" default:"5"`
	ShowHeader bool   `json:"show_header" default:"true"`
	ShowFooter bool   `json:"show_footer" default:"true"`
}

func newFileConfig() fileConfig {
	var cfg fileConfig
	defaults.SetDefaults(&cfg)
	return cfg
}

func (c *fileConfig) SwitchShowMode() {

	if !c.ShowHeader && !c.ShowFooter {
		c.ShowHeader = true
	} else if c.ShowHeader && !c.ShowFooter {
		c.ShowFooter = true
		c.ShowHeader = false
	} else if !c.ShowHeader && c.ShowFooter {
		c.ShowHeader = true
		c.ShowFooter = true
	} else {
		c.ShowHeader = false
		c.ShowFooter = false
	}
}

func (c *fileConfig) GetNodes() []string {
	return strings.Split(c.Nodes, ",")
}

func LoadFileConfig() tea.Msg {

	bf, err := os.ReadFile(SavePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return messages.LoadConfigResult{Error: nil}
		}
	}

	return messages.LoadConfigResult{Error: json.Unmarshal(bf, &G)}
}

func SaveToFile() tea.Msg {
	bytesData, err := json.Marshal(G)
	if err != nil {
		return err
	}

	return os.WriteFile(SavePath(), bytesData, 0644)
}

func SavePath() string {
	home, err := homedir.Dir()
	if err != nil {
		home = "."
	}

	return path.Join(home, ".go-v2ex.json")
}
