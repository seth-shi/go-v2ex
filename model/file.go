package model

import (
	"encoding/json"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/seth-shi/go-v2ex/v2/consts"
)

const (
	envProduction = "production"
)

type FileConfig struct {
	Token       string `json:"personal_access_token"`
	MyNodes     string `json:"my_nodes"`
	Timeout     uint   `json:"timeout"`
	ActiveTab   int    `json:"active_tab"`
	ShowMode    int    `json:"show_mode"`
	Env         string `json:"env"`
	ChooseAPIV2 bool   `json:"choose_api_v2"`
}

func NewDefaultFileConfig() *FileConfig {
	return &FileConfig{
		// NOTE: 增加默认秘钥, 方便用户快速使用, 用户以后还是要自己配置
		Token:       "35bbd155-df12-4778-9916-5dd59d967fef",
		MyNodes:     "share,create,qna,jobs,programmer,career,invest,ideas,hardware",
		Timeout:     5,
		ActiveTab:   0,
		ShowMode:    consts.ShowModeAll,
		Env:         envProduction,
		ChooseAPIV2: false,
	}
}

func (c *FileConfig) IsProductionEnv() bool {
	return strings.Contains(c.Env, "prod")
}
func (c *FileConfig) IsDevelopmentEnv() bool {
	return strings.Contains(c.Env, "dev")
}
func (c *FileConfig) IsMockEnv() bool {
	return strings.Contains(c.Env, "mock")
}

func (c *FileConfig) SwitchShowMode() int {
	c.ShowMode++
	if c.ShowMode > consts.ShowModeHideAll {
		c.ShowMode = consts.ShowModeAll
	}

	return c.ShowMode
}
func (c *FileConfig) GetShowModeText() string {
	var (
		m = map[int]string{
			consts.ShowModeAll:       "显示所有",
			consts.ShowModeHideLimit: "隐藏请求限制",
			consts.ShowModeHideAll:   "隐藏所有",
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
	return c.ShowMode != consts.ShowModeHideAll &&
		c.ShowMode != consts.ShowModeHideLimit
}

func SaveToFile(conf *FileConfig) error {
	bytesData, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), bytesData, 0644)
}

func ConfigPath() string {
	home, err := homedir.Dir()
	if err != nil {
		home = "."
	}

	return path.Join(home, ".go-v2ex.json")
}
