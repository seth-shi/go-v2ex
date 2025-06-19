package main

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
)

// these information will be collected when build, by `-ldflags "-X main.appVersion=0.1"`
const (
	defaultAppVersion = "0.0.0"
)

var (
	appVersion = defaultAppVersion
)

func init() {
	rebuildAppVersion()
}

func rebuildAppVersion() {
	if appVersion != defaultAppVersion {
		return
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	infoMap := lo.SliceToMap(
		info.Settings, func(item debug.BuildSetting) (string, string) {
			return item.Key, item.Value
		},
	)
	appVersion = formatBuildVersion(infoMap["vcs.revision"], infoMap["vcs.time"])
}

func formatBuildVersion(revision, vcsTime string) string {

	var prefix, suffix string
	if revision != "" && len(revision) >= 7 {
		suffix = revision[:7]
	}

	if vcsTime != "" {
		vcsCarbon := carbon.ParseByLayout(vcsTime, time.RFC3339)
		if !vcsCarbon.IsZero() {
			prefix = vcsCarbon.Layout("2006.01.02")
		}
	}
	return fmt.Sprintf("%s.%s", prefix, suffix)
}
