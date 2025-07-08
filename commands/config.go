package commands

import (
	"encoding/json"
	"errors"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seth-shi/go-v2ex/g"
	"github.com/seth-shi/go-v2ex/messages"
	"github.com/seth-shi/go-v2ex/model"
)

func LoadConfig() tea.Cmd {
	return func() tea.Msg {
		// 获取默认的配置, 或者重新读取的配置
		cfg := g.Config.Get()
		bf, err := os.ReadFile(model.ConfigPath())
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		err = json.Unmarshal(bf, &cfg)
		if err != nil {
			return err
		}

		// TODO 保存配置生效
		return messages.LoadConfigResult{Result: cfg}
	}
}
