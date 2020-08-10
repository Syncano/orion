package query

import (
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// UserMembershipManager represents User Membership manager.
type UserMembershipManager struct {
	*Factory
	*manager.Manager
}

// NewUserMembershipManager creates and returns new User Membership manager.
func (q *Factory) NewUserMembershipManager(c echo.Context) *UserMembershipManager {
	return &UserMembershipManager{Factory: q, Manager: manager.NewTenantManager(WrapContext(c), q.db)}
}

// Q outputs objects query.
func (m *UserMembershipManager) Q(o interface{}) *orm.Query {
	return m.Query(o)
}

// ForUserQ outputs objects query filtered by user.
func (m *UserMembershipManager) ForUserQ(user *models.User, o interface{}) *orm.Query {
	return m.Q(o).Where("user_id = ?", user.ID)
}

// ForGroupQ outputs objects query filtered by group.
func (m *UserMembershipManager) ForGroupQ(group *models.UserGroup, o interface{}) *orm.Query {
	return m.Q(o).Where("group_id = ?", group.ID)
}

// ForUserAndGroupQ outputs objects query filtered by user and group.
func (m *UserMembershipManager) ForUserAndGroupQ(o *models.UserMembership) *orm.Query {
	return m.Q(o).Where("user_id = ?", o.UserID).Where("group_id = ?", o.GroupID)
}
