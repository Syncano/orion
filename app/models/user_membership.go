package models

import (
	"fmt"
)

// UserMembership represents UserMembership model.
type UserMembership struct {
	tableName struct{} `pg:"?schema.users_membership,discard_unknown_columns"` // nolint

	ID      int
	UserID  int
	GroupID int

	User  *User
	Group *UserGroup
}

func (m *UserMembership) String() string {
	return fmt.Sprintf("UserMembership<ID=%d, UserID=%d, GroupID=%d>", m.ID, m.UserID, m.GroupID)
}

// VerboseName returns verbose name for model.
func (m *UserMembership) VerboseName() string {
	return "User Membership"
}
