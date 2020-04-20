package settings

import (
	"fmt"
	"os"
	"strings"
)

func GetCurrentLocationEnv(key string) string {
	return GetLocationEnv(Common.Location, key)
}

func GetCurrentLocationEnvDefault(key, defval string) string {
	return GetLocationEnvDefault(Common.Location, key, defval)
}

func GetLocationEnv(loc, key string) string {
	return os.Getenv(fmt.Sprintf("%s_%s", strings.ToUpper(loc), key))
}

func GetLocationEnvDefault(loc, key, defval string) string {
	val, ok := os.LookupEnv(fmt.Sprintf("%s_%s", strings.ToUpper(loc), key))
	if ok {
		return val
	}

	return defval
}
