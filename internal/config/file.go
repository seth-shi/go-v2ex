package config

import (
	"path"

	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/go-homedir"
	"github.com/seth-shi/go-v2ex/internal/consts"
)

const (
	envProduction  = "production"
	envDevelopment = "development"
	envMock        = "mock"
)

var (
	G = NewFileConfig()
)

type FileConfig struct {
	// NOTE: 增加默认秘钥, 方便用户快速使用, 用户以后还是要自己配置
	Token     string `json:"personal_access_token" default:"35bbd155-df12-4778-9916-5dd59d967fef"`
	MyNodes   string `json:"my_nodes" default:"share,create,qna,jobs,programmer,career,invest,ideas,hardware"`
	Timeout   uint   `json:"timeout" default:"5"`
	ActiveTab int    `json:"active_tab"`
	ShowMode  int    `json:"show_mode" default:"3"`
	Env       string `json:"env" default:"production"`
}

func NewFileConfig() *FileConfig {
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
	if c.ShowMode > consts.ShowModeHideAll {
		c.ShowMode = consts.ShowModeAll
	}
}
func (c *FileConfig) GetShowModeText() string {
	var (
		m = map[int]string{
			consts.ShowModeAll:       "显示所有",
			consts.ShowModeHideAll:   "隐藏所有",
			consts.ShowModeHideLimit: "不显示请求限制",
			consts.ShowModeHideHelp:  "隐藏帮助",
		}
	)

	return m[c.ShowMode]
}

func (c *FileConfig) ShowFooter() bool {
	return c.ShowMode != consts.ShowModeHideAll
}

func (c *FileConfig) ShowHelp() bool {
	return c.ShowMode == consts.ShowModeAll
}

func (c *FileConfig) ShowLimit() bool {
	return c.ShowMode == consts.ShowModeAll ||
		c.ShowMode == consts.ShowModeHideLimit
}

func SavePath() string {
	home, err := homedir.Dir()
	if err != nil {
		home = "."
	}

	return path.Join(home, ".go-v2ex.json")
}
