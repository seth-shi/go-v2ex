package routes

import (
	"github.com/seth-shi/go-v2ex/internal/ui/components/boss"
	"github.com/seth-shi/go-v2ex/internal/ui/components/detail"
	"github.com/seth-shi/go-v2ex/internal/ui/components/help"
	"github.com/seth-shi/go-v2ex/internal/ui/components/setting"
	"github.com/seth-shi/go-v2ex/internal/ui/components/splash"
	"github.com/seth-shi/go-v2ex/internal/ui/components/topics"
)

var (
	HelpModel       = help.New()
	SettingModel    = setting.New()
	TopicsModel     = topics.New()
	SplashModel     = splash.New()
	BossComingModel = boss.New()
	DetailModel     = detail.New()
)
