package query

import (
	"fmt"

	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// UserGroupManager represents User Group manager.
type UserGroupManager struct {
	*Factory
	*manager.LiveManager
}

// NewUserGroupManager creates and returns new User Group manager.
func (q *Factory) NewUserGroupManager(c echo.Context) *UserGroupManager {
	return &UserGroupManager{Factory: q, LiveManager: manager.NewLiveTenantManager(WrapContext(c), q.db)}
}

// Q outputs objects query.
func (m *UserGroupManager) Q(o interface{}) *orm.Query {
	return m.Query(o)
}

// ByIDQ outputs one object filtered by id.
func (m *UserGroupManager) ByIDQ(o *models.UserGroup) *orm.Query {
	return m.Q(o).Where("?TableAlias.id = ?", o.ID)
}

// ForUserQ outputs objects filtered by user.
func (m *UserGroupManager) ForUserQ(user *models.User, o interface{}) *orm.Query {
	return m.Q(o).
		Join("JOIN ?schema.users_membership AS m ON m.group_id = ?TableAlias.id AND m.user_id = ?", user.ID)
}

// ForUserByIDQ outputs one object filtered by user and id.
func (m *UserGroupManager) ForUserByIDQ(user *models.User, o *models.UserGroup) *orm.Query {
	return m.ForUserQ(user, o).Where("?TableAlias.id = ?", o.ID)
}

// OneByID outputs object filtered by id.
func (m *UserGroupManager) OneByID(o *models.UserGroup) error {
	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, m.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}
