package query

import (
	"fmt"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/storage"
)

// UserGroupManager represents User Group manager.
type UserGroupManager struct {
	*LiveManager
}

// NewUserGroupManager creates and returns new User Group manager.
func (q *Factory) NewUserGroupManager(c storage.DBContext) *UserGroupManager {
	return &UserGroupManager{LiveManager: q.NewLiveTenantManager(c)}
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
	return RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, m.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}
