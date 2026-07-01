package version

import (
	"runtime/debug"
	"time"
)

// Version of the app
var Version = "snapshot"

// GitCommit is the GIT commit revision
var GitCommit = "n/a"

// Built is the built date
var Built = "n/a"

func init() {
	if GitCommit != "snapshot" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	Version = info.Main.Version
	for _, kv := range info.Settings {
		if kv.Value == "" {
			continue
		}
		switch kv.Key {
		case "vcs.revision":
			GitCommit = kv.Value[:7]
		case "vcs.time":
			lastCommit, _ := time.Parse(time.RFC3339, kv.Value)
			Built = lastCommit.Format(time.RFC1123)
		}
	}
}
