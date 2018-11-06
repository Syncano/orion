package models

import (
	"fmt"
	"time"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
)

// Instance represents Instance (tenant) model.
type Instance struct {
	tableName struct{} `sql:"instances_instance" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID         int
	Name       string
	OwnerID    int
	Owner      *Admin
	SchemaName string
	Version    int
	Location   string

	CreatedAt     Time
	UpdatedAt     Time
	StoragePrefix string
	Config        JSON
	Description   string
	Metadata      JSON
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
func (m *Instance) BeforeUpdate(db orm.DB) error {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return nil
}

// AfterUpdate hook.
func (m *Instance) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *Instance) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}
