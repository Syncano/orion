package models

import (
	"fmt"

	"github.com/alexandrevicenzi/unchained"

	"github.com/Syncano/pkg-go/util"
)

// User represents User model.
type User struct {
	State

	tableName struct{} `pg:"?schema.users_user,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

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
