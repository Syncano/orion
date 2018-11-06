package models

import (
	"fmt"
)

// UserGroup represents User Group model.
type UserGroup struct {
	tableName struct{} `sql:"?schema.users_group" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID          int
	Name        string
	Label       string
	Description string
	CreatedAt   Time

	Users []*User `pg:"many2many:?schema.users_membership,joinFK:group_id"`
}

func (m *UserGroup) String() string {
	return fmt.Sprintf("UserGroup<ID=%d, Name=%q>", m.ID, m.Name)
}

// VerboseName returns verbose name for model.
func (m *UserGroup) VerboseName() string {
	return "User Group"
}
