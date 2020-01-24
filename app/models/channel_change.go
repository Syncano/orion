package models

import (
	"fmt"
	"time"
)

const (
	changeListMaxSize = 100
	changeTTL         = 24 * time.Hour
	changeTrimmedTTL  = 1 * time.Hour

	changeEventlogListMaxSize = 1000
	changeEventlogTrimmedTTL  = 5 * time.Minute
)

// ChangeAction enum.
const (
	ChangeActionCustom int = iota
	ChangeActionCreate
	ChangeActionUpdate
	ChangeActionDelete
)

// ChangeAction to string map.
var ChangeAction = map[int]string{
	ChangeActionCustom: "custom",
	ChangeActionCreate: "create",
	ChangeActionUpdate: "update",
	ChangeActionDelete: "delete",
}

// Change represents channel change redis model.
type Change struct {
	ID        int
	CreatedAt time.Time
	Action    int
	Author    map[string]interface{} `default:"{}"`
	Metadata  map[string]interface{} `default:"{}"`
	Payload   map[string]interface{} `default:"{}"`
}

// VerboseName returns verbose name for model.
func (m *Change) VerboseName() string {
	return "Channel Change"
}

// ActionString returns verbose value for action.
func (m *Change) ActionString() string {
	return ChangeAction[m.Action]
}

func (m *Change) Key(args map[string]interface{}) string {
	return fmt.Sprintf("%d:rdb:Change", args["instance"].(*Instance).ID)
}

func (m *Change) ListArgs(args map[string]interface{}) string {
	return fmt.Sprintf("%d:%v", args["channel"].(*Channel).ID, args["room"])
}

func (m *Change) ListMaxSize(args map[string]interface{}) int {
	if args["channel"].(*Channel).Name == ChannelEventlogName {
		return changeEventlogListMaxSize
	}

	return changeListMaxSize
}

func (m *Change) TTL(args map[string]interface{}) time.Duration {
	return changeTTL
}

func (m *Change) TrimmedTTL(args map[string]interface{}) time.Duration {
	if args["channel"].(*Channel).Name == ChannelEventlogName {
		return changeEventlogTrimmedTTL
	}

	return changeTrimmedTTL
}
