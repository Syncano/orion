package models

import (
	"fmt"
	"time"

	"github.com/go-pg/pg/orm"
)

// TriggerSignal enum.
const (
	TriggerSignalCreate = "create"
	TriggerSignalUpdate = "update"
	TriggerSignalDelete = "delete"
)

// Trigger represents trigger model.
type Trigger struct {
	tableName struct{} `sql:"?schema.triggers_trigger" pg:",discard_unknown_columns"` // nolint

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
func (m *Trigger) BeforeUpdate(db orm.DB) error {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return nil
}
