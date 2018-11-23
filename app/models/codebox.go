package models

import (
	"fmt"
	"time"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
)

// Codebox represents codebox model.
type Codebox struct {
	tableName struct{} `sql:"?schema.codeboxes_codebox" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID          int
	Description string
	Label       string
	RuntimeName string
	Checksum    string
	Path        string
	Config      JSON
	CreatedAt   Time
	UpdatedAt   Time

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
func (m *Codebox) BeforeUpdate(db orm.DB) error {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return nil
}

// AfterUpdate hook.
func (m *Codebox) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *Codebox) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}
