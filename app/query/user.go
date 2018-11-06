package query

import (
	"fmt"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
)

// UserManager represents User manager.
type UserManager struct {
	*LiveManager
}

// NewUserManager creates and returns new User manager.
func NewUserManager(c storage.DBContext) *UserManager {
	return &UserManager{LiveManager: NewLiveTenantManager(c)}
}

// Q outputs objects query.
func (mgr *UserManager) Q(class *models.Class, o interface{}) *orm.Query {
	return mgr.Query(o).Column("user.*", "Profile").
		Where("profile._klass_id = ?", class.ID)
}

// ByIDQ outputs one object that entity has access to filtered by id.
func (mgr *UserManager) ByIDQ(class *models.Class, o *models.User) *orm.Query {
	return mgr.Q(class, o).Where("?TableAlias.id = ?", o.ID)
}

// ForGroupQ outputs objects that entity has access to filtered by group.
func (mgr *UserManager) ForGroupQ(class *models.Class, group *models.UserGroup, o interface{}) *orm.Query {
	return mgr.Q(class, o).
		Join("JOIN ?schema.users_membership AS m ON m.user_id = ?TableAlias.id AND m.group_id = ?", group.ID)
}

// ForGroupByIDQ outputs one object that entity has access to filtered by group and id.
func (mgr *UserManager) ForGroupByIDQ(class *models.Class, group *models.UserGroup, o *models.User) *orm.Query {
	return mgr.ForGroupQ(class, group, o).Where("?TableAlias.id = ?", o.ID)
}

// OneByID outputs object filtered by id.
func (mgr *UserManager) OneByID(o *models.User) error {
	return RequireOne(
		cache.SimpleModelCache(mgr.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, mgr.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}

// OneByName outputs object filtered by name.
func (mgr *UserManager) OneByName(o *models.User) error {
	return RequireOne(
		cache.SimpleModelCache(mgr.DB(), o, fmt.Sprintf("u=%s", o.Username), func() (interface{}, error) {
			return o, mgr.Query(o).Where("username = ?", o.Username).Select()
		}),
	)
}

// OneByKey outputs object filtered by key. Doesn't fetch profile nor groups.
func (mgr *UserManager) OneByKey(o *models.User) error {
	return RequireOne(
		cache.SimpleModelCache(mgr.DB(), o, fmt.Sprintf("k=%s", o.Key), func() (interface{}, error) {
			return o, mgr.Query(o).Where("key = ?", o.Key).Select()
		}),
	)
}

// FetchData outputs object filtered by name.
func (mgr *UserManager) FetchData(class *models.Class, o *models.User) error {
	return mgr.Query(o).Column("_", "Profile", "Groups").
		Where("profile._klass_id = ?", class.ID).WherePK().Select()
}

// CountEstimate returns count estimate for users list.
func (mgr *UserManager) CountEstimate() (int, error) {
	return CountEstimate(mgr.DB(), mgr.Query((*models.User)(nil)), settings.API.DataObjectEstimateThreshold)
}
