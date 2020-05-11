package util

import (
	"fmt"
	"os"
	"strings"
)

func GetPrefixEnv(prefix, key string) string {
	return os.Getenv(fmt.Sprintf("%s_%s", strings.ToUpper(prefix), key))
}

func GetPrefixEnvDefault(prefix, key, defval string) string {
	val, ok := os.LookupEnv(fmt.Sprintf("%s_%s", strings.ToUpper(prefix), key))
	if ok {
		return val
	}

	return defval
}
