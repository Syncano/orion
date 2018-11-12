package models

import (
	"fmt"
	"time"
)

const (
	codeboxListMaxSize = 100
	codeboxTTL         = 24 * time.Hour
	codeboxTrimmedTTL  = 5 * time.Minute
)

// TraceStatus enum.
const (
	TraceStatusBlocked    = "blocked"
	TraceStatusPending    = "pending"
	TraceStatusSuccess    = "success"
	TraceStatusFailure    = "failure"
	TraceStatusTimeout    = "timeout"
	TraceStatusProcessing = "processing"
)

// SocketTrace represents socket trace redis model.
type SocketTrace struct {
	ID         int
	Status     string `default:"pending"`
	ExecutedAt time.Time
	Duration   int
	Result     map[string]interface{} `default:"{}"`
	Meta       map[string]interface{} `default:"{}"`
	Args       map[string]interface{} `default:"{}"`
}

// VerboseName returns verbose name for model.
func (m *SocketTrace) VerboseName() string {
	return "Socket Trace"
}

// Key ...
func (m *SocketTrace) Key(args map[string]interface{}) string {
	return fmt.Sprintf("%v:rdb:SocketEndpointTrace", args["instance_id"])
}

// ListArgs ...
func (m *SocketTrace) ListArgs(args map[string]interface{}) string {
	return fmt.Sprintf("%v", args["socket_endpoint_id"])
}

// ListMaxSize ...
func (m *SocketTrace) ListMaxSize() int {
	return codeboxListMaxSize
}

// TTL ...
func (m *SocketTrace) TTL() time.Duration {
	return codeboxTTL
}

// TrimmedTTL ...
func (m *SocketTrace) TrimmedTTL() time.Duration {
	return codeboxTrimmedTTL
}
