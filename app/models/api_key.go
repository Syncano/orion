package models

import (
	"fmt"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
)

// APIKey represents API Key model.
type APIKey struct {
	State
	tableName struct{} `sql:"apikeys_apikey"` // nolint

	IsLive bool `sql:"_is_live"`

	ID          int
	Key         string
	InstanceID  int
	Instance    *Instance
	Options     *Hstore
	CreatedAt   *Time
	Description string
}

func (m *APIKey) String() string {
	return fmt.Sprintf("APIKey<ID=%d Instance=%d>", m.ID, m.InstanceID)
}

// VerboseName returns verbose name for model.
func (m *APIKey) VerboseName() string {
	return "API Key"
}

// AfterUpdate hook.
func (m *APIKey) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *APIKey) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}
