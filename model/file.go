package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/seth-shi/go-v2ex/v2/consts"
)

const (
	envProduction                = "production"
	legacyDefaultTokenSHA256Hash = "8b4e43bc20cd911bfc63695c73738d15c0cd624a15ded7e15f8afd7fa753d8ef"
)

type FileConfig struct {
	Token          string `json:"personal_access_token"`
	MyNodes        string `json:"my_node_keys"`
	Timeout        uint   `json:"timeout"`
	ActiveTab      int    `json:"active_tab"`
	ShowMode       int    `json:"show_mode"`
	Env            string `json:"env"`
	OnboardingDone bool   `json:"onboarding_done"`
}

func NewDefaultFileConfig() *FileConfig {
	return &FileConfig{
		Token:     "",
		MyNodes:   "",
		Timeout:   5,
		ActiveTab: 0,
		ShowMode:  consts.ShowModeAll,
		Env:       envProduction,
	}
}

// ClearLegacyDefaultToken removes only the credential bundled by older
// releases. Personal tokens remain untouched.
func (c *FileConfig) ClearLegacyDefaultToken() bool {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(c.Token))))
	if tokenHash != legacyDefaultTokenSHA256Hash {
		return false
	}
	c.Token = ""
	return true
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

	configPath := ConfigPath()
	if err = os.WriteFile(configPath, bytesData, 0600); err != nil {
		return err
	}
	// WriteFile keeps the mode of an existing file, so enforce private access
	// for configurations created by older releases as well.
	return os.Chmod(configPath, 0600)
}

func ConfigPath() string {
	home, err := homedir.Dir()
	if err != nil {
		home = "."
	}

	return path.Join(home, ".go-v2ex.json")
}
