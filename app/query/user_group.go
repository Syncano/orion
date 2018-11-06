package query

import (
	"fmt"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
)

// UserGroupManager represents User Group manager.
type UserGroupManager struct {
	*LiveManager
}

// NewUserGroupManager creates and returns new User Group manager.
func NewUserGroupManager(c storage.DBContext) *UserGroupManager {
	return &UserGroupManager{LiveManager: NewLiveTenantManager(c)}
}

// Q outputs objects query.
func (mgr *UserGroupManager) Q(o interface{}) *orm.Query {
	return mgr.Query(o)
}

// ByIDQ outputs one object filtered by id.
func (mgr *UserGroupManager) ByIDQ(o *models.UserGroup) *orm.Query {
	return mgr.Q(o).Where("?TableAlias.id = ?", o.ID)
}

// ForUserQ outputs objects filtered by user.
func (mgr *UserGroupManager) ForUserQ(user *models.User, o interface{}) *orm.Query {
	return mgr.Q(o).
		Join("JOIN ?schema.users_membership AS m ON m.group_id = ?TableAlias.id AND m.user_id = ?", user.ID)
}

// ForUserByIDQ outputs one object filtered by user and id.
func (mgr *UserGroupManager) ForUserByIDQ(user *models.User, o *models.UserGroup) *orm.Query {
	return mgr.ForUserQ(user, o).Where("?TableAlias.id = ?", o.ID)
}

// OneByID outputs object filtered by id.
func (mgr *UserGroupManager) OneByID(o *models.UserGroup) error {
	return RequireOne(
		cache.SimpleModelCache(mgr.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, mgr.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}
