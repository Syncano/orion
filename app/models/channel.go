package models

import (
	"context"
	"fmt"
	"time"

	"github.com/Syncano/pkg-go/v2/database/fields"
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
	tableName struct{} `pg:"?schema.channels_channel,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

	ID          int
	Name        string
	Type        int
	Description string

	CreatedAt fields.Time
	UpdatedAt fields.Time
}

func (m *Channel) String() string {
	return fmt.Sprintf("Channel<ID=%d Name=%q>", m.ID, m.Name)
}

// VerboseName returns verbose name for model.
func (m *Channel) VerboseName() string {
	return "Channel"
}

// BeforeUpdate hook.
func (m *Channel) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return ctx, nil
}
