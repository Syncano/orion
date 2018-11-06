package models

import (
	"fmt"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
)

// AdminInstanceRole represents AdminInstanceRole model.
type AdminInstanceRole struct {
	tableName struct{} `sql:"admins_admininstancerole"` // nolint

	ID         int
	AdminID    int
	Admin      *Admin
	InstanceID int
	Instance   *Instance
}

func (m *AdminInstanceRole) String() string {
	return fmt.Sprintf("AdminInstanceRole<Admin=%d Instance=%d>", m.AdminID, m.InstanceID)
}

// VerboseName returns verbose name for model.
func (m *AdminInstanceRole) VerboseName() string {
	return "Admin Instance Role"
}

// AfterUpdate hook.
func (m *AdminInstanceRole) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *AdminInstanceRole) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}
