package settings

import (
	"github.com/Syncano/pkg-go/util"
)

func GetCurrentLocationEnv(key string) string {
	return GetLocationEnv(Common.Location, key)
}

func GetCurrentLocationEnvDefault(key, defval string) string {
	return GetLocationEnvDefault(Common.Location, key, defval)
}

func GetLocationEnv(loc, key string) string {
	return util.GetPrefixEnv(loc, key)
}

func GetLocationEnvDefault(loc, key, defval string) string {
	return util.GetPrefixEnvDefault(loc, key, defval)
}
