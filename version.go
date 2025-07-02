package main

import (
	"runtime/debug"
)

// these information will be collected when build, by `-ldflags "-X main.appVersion=0.0.1"`
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

	if info.Main.Version == "(devel)" {
		return
	}

	appVersion = info.Main.Version
}
