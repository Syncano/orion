package version

import (
	"time"

	"github.com/blang/semver"
)

// GitSHA, buildtimeStr and versionStr are set during a build.
var (
	Buildtime    time.Time
	Current      semver.Version
	GitSHA       = "dev"
	buildtimeStr string
	versionStr   = "0.0.1"
)

func mustParseBuildtime(buildtimeStr string) time.Time {
	if buildtimeStr == "" {
		return time.Now()
	}
	b, err := time.Parse("2006-01-02T15:04", buildtimeStr)
	if err != nil {
		panic(err)
	}
	return b
}

func init() {
	Current = semver.MustParse(versionStr)
	Buildtime = mustParseBuildtime(buildtimeStr)
}
