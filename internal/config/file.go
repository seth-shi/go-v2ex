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

var (
	G = newFileConfig()
)

type fileConfig struct {
	// NOTE: 增加默认秘钥, 方便用户快速使用, 用户以后还是要自己配置
	Token      string `json:"personal_access_token" default:"35bbd155-df12-4778-9916-5dd59d967fef"`
	Nodes      string `json:"nodes" default:"latest,hot,qna,all4all,programmer,jobs,share,apple,create,macos,career,pointless"`
	Timeout    uint   `json:"timeout" default:"5"`
	ShowFooter bool   `json:"show_footer" default:"true"`
	ActiveTab  int    `json:"active_tab"`
}

func newFileConfig() fileConfig {
	var cfg fileConfig
	defaults.SetDefaults(&cfg)
	return cfg
}

func (c *fileConfig) SwitchShowMode() {
	c.ShowFooter = !c.ShowFooter
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

func SaveToFile(title string) tea.Cmd {
	return func() tea.Msg {
		bytesData, err := json.Marshal(G)
		if err != nil {
			return err
		}

		if err = os.WriteFile(SavePath(), bytesData, 0644); err != nil {
			return err
		}

		if title == "" {
			return nil
		}

		return messages.ShowAutoTipsRequest{Text: title}
	}
}

func SavePath() string {
	home, err := homedir.Dir()
	if err != nil {
		home = "."
	}

	return path.Join(home, ".go-v2ex.json")
}
