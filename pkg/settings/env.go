package settings

import (
	"fmt"
	"os"
	"strings"
)

func GetCurrentLocationEnv(key string) string {
	return GetLocationEnv(Common.Location, key)
}

func GetLocationEnv(loc, key string) string {
	return os.Getenv(fmt.Sprintf("%s_%s", strings.ToUpper(loc), key))
}
