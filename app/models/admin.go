package models

import (
	"fmt"

	"github.com/alexandrevicenzi/unchained"

	"github.com/Syncano/pkg-go/database/fields"
	"github.com/Syncano/pkg-go/util"
)

// Admin represents Admin model.
type Admin struct {
	State
	tableName struct{} `pg:"admins_admin,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

	ID        int
	Email     string
	Key       string
	Password  string
	FirstName string
	LastName  string
	IsStaff   bool
	IsActive  bool

	CreatedAt  fields.Time
	LastLogin  fields.Time
	LastAccess fields.Time
	NoticedAt  fields.Time
	Metadata   fields.JSON
}

func (m *Admin) String() string {
	return fmt.Sprintf("Admin<ID=%d Email=%q>", m.ID, m.Email)
}

// VerboseName returns verbose name for model.
func (m *Admin) VerboseName() string {
	return "Admin"
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
