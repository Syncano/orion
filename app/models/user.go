package models

import (
	"fmt"

	"github.com/alexandrevicenzi/unchained"
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/util"
)

// User represents User model.
type User struct {
	State

	tableName struct{} `sql:"?schema.users_user" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID        int
	Username  string
	Password  string
	Key       string
	CreatedAt Time

	Profile *DataObject  `pg:"fk:owner_id" msgpack:"-"`
	Groups  []*UserGroup `pg:"many2many:?schema.users_membership,joinFK:group_id" msgpack:"-"`
}

func (m *User) String() string {
	return fmt.Sprintf("User<ID=%d, Username=%q>", m.ID, m.Username)
}

// VerboseName returns verbose name for model.
func (m *User) VerboseName() string {
	return "User"
}

// AfterUpdate hook.
func (m *User) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *User) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// SetPassword sets encrypted password on object.
func (m *User) SetPassword(password string) {
	m.Password = util.MakePassword(password)
}

// CheckPassword checks if provided password is correct for object.
func (m *User) CheckPassword(pwd string) bool {
	return util.VerifyPassword(pwd, m.Password)
}

// IsPasswordUsable checks if current password is usable.
func (m *User) IsPasswordUsable() bool {
	return unchained.IsPasswordUsable(m.Password)
}
