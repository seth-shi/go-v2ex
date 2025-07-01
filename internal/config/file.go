package config

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/go-homedir"
	"github.com/samber/lo"
	"github.com/seth-shi/go-v2ex/internal/consts"
)

const (
	envProduction  = "production"
	envDevelopment = "development"
	envMock        = "mock"
)

var (
	G = newFileConfig()
)

type FileConfig struct {
	// NOTE: 增加默认秘钥, 方便用户快速使用, 用户以后还是要自己配置
	Token     string `json:"personal_access_token" default:"35bbd155-df12-4778-9916-5dd59d967fef"`
	MyNodes   string `json:"my_nodes" default:"share,create,qna,jobs,programmer,career,invest,ideas,hardware"`
	Timeout   uint   `json:"timeout" default:"5"`
	ActiveTab int    `json:"active_tab"`
	ShowMode  int    `json:"show_mode" default:"4"`
	Env       string `json:"env" default:"production"`
}

func newFileConfig() *FileConfig {
	var cfg FileConfig
	defaults.SetDefaults(&cfg)
	return &cfg
}

func (c *FileConfig) IsProductionEnv() bool {
	return c.Env == envProduction
}
func (c *FileConfig) IsDevelopmentEnv() bool {
	return c.Env == envDevelopment
}
func (c *FileConfig) IsMockEnv() bool {
	return c.Env == envMock
}

func (c *FileConfig) SwitchShowMode() {
	c.ShowMode++
	if c.ShowMode > consts.ShowModeAll {
		c.ShowMode = consts.ShowModeHidden
	}
}
func (c *FileConfig) GetShowModeText() string {
	var (
		m = map[int]string{
			consts.ShowModeHidden:                "隐藏所有底部",
			consts.ShowModeLeftAndRight:          "不显示请求限制",
			consts.ShowModeLeftAndRightWithLimit: "显示左侧和右侧+请求限制量",
			consts.ShowModeAll:                   "显示所有",
		}
	)

	return m[c.ShowMode]
}

func (c *FileConfig) ShowFooter() bool {
	return c.ShowMode != consts.ShowModeHidden
}

func (c *FileConfig) ShowHelp() bool {
	return c.ShowMode == consts.ShowModeAll
}

func (c *FileConfig) ShowLimit() bool {
	return c.ShowMode == consts.ShowModeLeftAndRightWithLimit ||
		c.ShowMode == consts.ShowModeAll
}

func LoadFileConfig() (FileConfig, error) {

	bf, err := os.ReadFile(SavePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return lo.FromPtr(G), nil
		}
	}

	err = json.Unmarshal(bf, &G)
	return lo.FromPtr(G), err
}

func SaveToFile() error {
	bytesData, err := json.Marshal(G)
	if err != nil {
		return err
	}

	if err = os.WriteFile(SavePath(), bytesData, 0644); err != nil {
		return err
	}

	return nil
}

func SavePath() string {
	home, err := homedir.Dir()
	if err != nil {
		home = "."
	}

	return path.Join(home, ".go-v2ex.json")
}
