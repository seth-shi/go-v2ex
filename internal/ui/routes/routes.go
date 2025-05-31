package routes

import (
	"github.com/seth-shi/go-v2ex/internal/ui/components/help"
	"github.com/seth-shi/go-v2ex/internal/ui/components/setting"
	"github.com/seth-shi/go-v2ex/internal/ui/components/splash"
	"github.com/seth-shi/go-v2ex/internal/ui/components/topics"
)

var (
	HelpModel    help.Model
	SettingModel setting.Model
	TopicsModel  topics.Model
	SplashModel  splash.Model
)

func init() {
	HelpModel = help.New()
	SettingModel = setting.New()
	TopicsModel = topics.New()
	SplashModel = splash.New()
}
