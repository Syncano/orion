package validators

import (
	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
)

type UserAuthForm struct {
	Username string `form:"username" validate:"required"`
	Password string `form:"password" validate:"required"`
}

type UserInGroupForm struct {
	UserQ       *orm.Query
	MembershipQ *orm.Query
	// Validate:
	// sql_select: make sure UserQ.Where(id=this_value).Select() returns no error
	// sql_notexists: make sure ! MembershipQ.Where(user_id=this_value).Exists()
	User int `form:"user" validate:"required,sql_select,sql_notexists=user_id MembershipQ"`
}

func (f *UserInGroupForm) Bind(m *models.UserMembership) {
	m.UserID = f.User
}

type GroupInUserForm struct {
	GroupQ      *orm.Query
	MembershipQ *orm.Query
	// sql_select: make sure UserQ.Where(id=this_value).Select() returns no error
	// sql_notexists: make sure ! MembershipQ.Where(group_id=this_value).Exists()
	Group int `form:"group" validate:"required,sql_select,sql_notexists=group_id MembershipQ"`
}

func (f *GroupInUserForm) Bind(m *models.UserMembership) {
	m.GroupID = f.Group
}
