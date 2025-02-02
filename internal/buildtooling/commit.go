package buildtooling

import "runtime/debug"

type BuildInfo struct {
	CleanTree bool
	Commit    string
	Time      string
}

func getBuildInfo() BuildInfo {
	var result BuildInfo
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.modified":
				result.CleanTree = setting.Value == "false"
			case "vcs.revision":
				result.Commit = setting.Value
			case "vcs.time":
				result.Time = setting.Value
			}
		}
	}
	return result
}

var Build = getBuildInfo()
