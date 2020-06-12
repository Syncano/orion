package models

import (
	"context"
	"fmt"
	"time"

	"github.com/Syncano/pkg-go/database/fields"
)

// Instance represents Instance (tenant) model.
type Instance struct {
	tableName struct{} `pg:"instances_instance,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

	ID         int
	Name       string
	OwnerID    int
	Owner      *Admin
	SchemaName string
	Version    int
	Location   string

	CreatedAt     fields.Time
	UpdatedAt     fields.Time
	StoragePrefix string
	Config        fields.JSON
	Description   string
	Metadata      fields.JSON
	Domains       []string `pg:",array"`
}

func (m *Instance) String() string {
	return fmt.Sprintf("Instance<ID=%d Name=%q>", m.ID, m.Name)
}

// VerboseName returns verbose name for model.
func (m *Instance) VerboseName() string {
	return "Instance"
}

// BeforeUpdate hook.
func (m *Instance) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return ctx, nil
}
