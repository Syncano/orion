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
	Weight     int                    `default:"1"`
	Result     map[string]interface{} `default:"{}"`
	Meta       map[string]interface{} `default:"{}"`
	Args       map[string]interface{} `default:"{}"`
}

// VerboseName returns verbose name for model.
func (m *SocketTrace) VerboseName() string {
	return "Socket Trace"
}

func (m *SocketTrace) Key(args map[string]interface{}) string {
	return fmt.Sprintf("%d:rdb:SocketEndpointTrace", args["instance"].(*Instance).ID)
}

func (m *SocketTrace) ListArgs(args map[string]interface{}) string {
	return fmt.Sprintf("%d", args["socket_endpoint"].(*SocketEndpoint).ID)
}

func (m *SocketTrace) ListMaxSize(args map[string]interface{}) int {
	return codeboxListMaxSize
}

func (m *SocketTrace) TTL(args map[string]interface{}) time.Duration {
	return codeboxTTL
}

func (m *SocketTrace) TrimmedTTL(args map[string]interface{}) time.Duration {
	return codeboxTrimmedTTL
}
