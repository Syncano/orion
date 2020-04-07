package models

import (
	"database/sql"
	"fmt"
)

// Event represents billing event model.
type Event struct {
	tableName struct{} `pg:"billing_event,discard_unknown_columns"` // nolint

	ID         int
	ExternalID string
	Type       string
	Livemode   bool
	Message    JSON
	Valid      sql.NullBool
	CreatedAt  Time
}

func (m *Event) String() string {
	return fmt.Sprintf("Event<ID=%d ExternalID=%q>", m.ID, m.ExternalID)
}

// VerboseName returns verbose name for model.
func (m *Event) VerboseName() string {
	return "Event"
}
