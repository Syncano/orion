package models

import (
	"fmt"

	"github.com/Syncano/pkg-go/database/fields"
)

// UserGroup represents User Group model.
type UserGroup struct {
	tableName struct{} `pg:"?schema.users_group,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

	ID          int
	Name        string
	Label       string
	Description string
	CreatedAt   fields.Time

	Users []*User `pg:"many2many:?schema.users_membership,joinFK:group_id"`
}

func (m *UserGroup) String() string {
	return fmt.Sprintf("UserGroup<ID=%d, Name=%q>", m.ID, m.Name)
}

// VerboseName returns verbose name for model.
func (m *UserGroup) VerboseName() string {
	return "User Group"
}
