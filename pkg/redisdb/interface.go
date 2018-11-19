package redisdb

import (
	"time"
)

// Modeler represents Redis Model interfaces.
//go:generate mockery -name Modeler
type Modeler interface {
	// Key returns base object/list key. Object may be nil.
	Key(args map[string]interface{}) string
	// ListArgs returns list key suffix. Object may be nil.
	ListArgs(args map[string]interface{}) string

	// ListMaxSize returns list max size properties.
	ListMaxSize(args map[string]interface{}) int
	// TTL returns list ttl.
	TTL(args map[string]interface{}) time.Duration
	// TrimmedTTL returns list trimmed ttl - ttl of objects that are no longer a part of a list.
	TrimmedTTL(args map[string]interface{}) time.Duration
}
