package models

import (
	"context"
	"fmt"
	"time"

	"github.com/Syncano/pkg-go/database/fields"
)

// Codebox represents codebox model.
type Codebox struct {
	tableName struct{} `pg:"?schema.codeboxes_codebox,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

	ID          int
	Description string
	Label       string
	RuntimeName string
	Checksum    string
	Path        string
	Config      fields.JSON
	CreatedAt   fields.Time
	UpdatedAt   fields.Time

	SocketID int
	Socket   *Socket
}

func (m *Codebox) String() string {
	return fmt.Sprintf("Codebox<ID=%d, Label=%q>", m.ID, m.Label)
}

// VerboseName returns verbose name for model.
func (m *Codebox) VerboseName() string {
	return "Codebox"
}

// BeforeUpdate hook.
func (m *Codebox) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return ctx, nil
}
