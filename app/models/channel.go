package models

import (
	"fmt"
	"time"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
)

// Channel constants.
const (
	ChannelDefaultName  = "default"
	ChannelEventlogName = "eventlog"
)

// ChannelType enum.
const (
	ChannelTypeDefault int = iota
	ChannelTypeSeparateRooms
)

// ChannelType to string map.
var ChannelType = map[int]string{
	ChannelTypeDefault:       "default",
	ChannelTypeSeparateRooms: "separate_rooms",
}

// Channel represents Channel model.
type Channel struct {
	tableName struct{} `sql:"?schema.channels_channel" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID          int
	Name        string
	Type        int
	Description string

	CreatedAt *Time
	UpdatedAt *Time
}

func (m *Channel) String() string {
	return fmt.Sprintf("Channel<ID=%d Name=%q>", m.ID, m.Name)
}

// VerboseName returns verbose name for model.
func (m *Channel) VerboseName() string {
	return "Channel"
}

// AfterUpdate hook.
func (m *Channel) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *Channel) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// BeforeUpdate hook.
func (m *Channel) BeforeUpdate(db orm.DB) error {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return nil
}
