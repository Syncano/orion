package models

import (
	"context"
	"fmt"
	"time"
)

// TriggerSignal enum.
const (
	TriggerSignalCreate = "create"
	TriggerSignalUpdate = "update"
	TriggerSignalDelete = "delete"
)

// Trigger represents trigger model.
type Trigger struct {
	tableName struct{} `pg:"?schema.triggers_trigger,discard_unknown_columns"` // nolint

	ID          int
	Description string
	Label       string
	CreatedAt   Time
	UpdatedAt   Time
	Event       Hstore
	Signals     []string

	CodeboxID int
	Codebox   *Codebox
	SocketID  int
	Socket    *Socket
}

func (m *Trigger) String() string {
	return fmt.Sprintf("Trigger<ID=%d, Label=%q>", m.ID, m.Label)
}

// VerboseName returns verbose name for model.
func (m *Trigger) VerboseName() string {
	return "Trigger"
}

// BeforeUpdate hook.
func (m *Trigger) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return ctx, nil
}
