package models

import (
	"fmt"

	"github.com/alexandrevicenzi/unchained"
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/util"
)

// Admin represents Admin model.
type Admin struct {
	State
	tableName struct{} `sql:"admins_admin" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID        int
	Email     string
	Key       string
	Password  string
	FirstName string
	LastName  string
	IsStaff   bool
	IsActive  bool

	CreatedAt  *Time
	LastLogin  *Time
	LastAccess *Time
	NoticedAt  *Time
	Metadata   *JSON
}

func (m *Admin) String() string {
	return fmt.Sprintf("Admin<ID=%d Email=%q>", m.ID, m.Email)
}

// VerboseName returns verbose name for model.
func (m *Admin) VerboseName() string {
	return "Admin"
}

// AfterUpdate hook.
func (m *Admin) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *Admin) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// SetPassword sets encrypted password on object.
func (m *Admin) SetPassword(password string) {
	m.Password = util.MakePassword(password)
}

// CheckPassword checks if provided password is correct for object.
func (m *Admin) CheckPassword(pwd string) bool {
	return util.VerifyPassword(pwd, m.Password)
}

// IsPasswordUsable checks if current password is usable.
func (m *Admin) IsPasswordUsable() bool {
	return unchained.IsPasswordUsable(m.Password)
}
