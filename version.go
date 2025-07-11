package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
)

// these information will be collected when build, by `-ldflags "-X main.appVersion=0.0.1"`
const (
	defaultAppVersion = "0.0.0"
	textLogo          = `
 _____ ____        _     ____  ________  _
/  __//  _ \      / \ |\/_   \/  __/\  \//
| |  _| / \|_____ | | // /   /|  \   \  / 
| |_//| \_/|\____\| \// /   /_|  /_  /  \ 
\____\\____/      \__/  \____/\____\/__/\\
                                          
`
)

var (
	appVersion = defaultAppVersion
)

func init() {
	rebuildAppVersion()

	var (
		versionFlag = flag.Bool("version", false, "显示版本号")
		msg         = fmt.Sprintf("%s \nversion %s", textLogo, appVersion)
	)
	flag.Parse()

	if *versionFlag {
		fmt.Println(msg)
		os.Exit(0)
	}
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
