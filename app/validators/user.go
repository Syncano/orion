package validators

import (
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/app/models"
)

// UserAuthForm ...
type UserAuthForm struct {
	Username string `form:"username" validate:"required"`
	Password string `form:"password" validate:"required"`
}

// UserInGroupForm ...
type UserInGroupForm struct {
	UserQ       *orm.Query
	MembershipQ *orm.Query
	User        int `form:"user" validate:"required,sql_select,sql_notexists=user_id MembershipQ"`
}

// Bind ...
func (f *UserInGroupForm) Bind(m *models.UserMembership) {
	m.UserID = f.User
}

// GroupInUserForm ...
type GroupInUserForm struct {
	GroupQ      *orm.Query
	MembershipQ *orm.Query
	Group       int `form:"group" validate:"required,sql_select,sql_notexists=group_id MembershipQ"`
}

// Bind ...
func (f *GroupInUserForm) Bind(m *models.UserMembership) {
	m.GroupID = f.Group
}
