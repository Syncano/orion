package redisdb

import (
	"time"
)

// Modeler represents Redis Model interfaces.
//go:generate mockery -name Modeler
type Modeler interface {
	Key(args map[string]interface{}) string
	ListArgs(args map[string]interface{}) string
	ListMaxSize() int
	TTL() time.Duration
	TrimmedTTL() time.Duration
}
