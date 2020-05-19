package query

import (
	"fmt"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/pkg/storage"
)

// UserManager represents User manager.
type UserManager struct {
	*LiveManager
}

// NewUserManager creates and returns new User manager.
func (q *Factory) NewUserManager(c storage.DBContext) *UserManager {
	return &UserManager{LiveManager: q.NewLiveTenantManager(c)}
}

// Q outputs objects query.
func (m *UserManager) Q(class *models.Class, o interface{}) *orm.Query {
	return m.Query(o).Column("user.*").Relation("Profile").
		Where("profile._klass_id = ?", class.ID)
}

// ByIDQ outputs one object that entity has access to filtered by id.
func (m *UserManager) ByIDQ(class *models.Class, o *models.User) *orm.Query {
	return m.Q(class, o).Where("?TableAlias.id = ?", o.ID)
}

// ForGroupQ outputs objects that entity has access to filtered by group.
func (m *UserManager) ForGroupQ(class *models.Class, group *models.UserGroup, o interface{}) *orm.Query {
	return m.Q(class, o).
		Join("JOIN ?schema.users_membership AS m ON m.user_id = ?TableAlias.id AND m.group_id = ?", group.ID)
}

// ForGroupByIDQ outputs one object that entity has access to filtered by group and id.
func (m *UserManager) ForGroupByIDQ(class *models.Class, group *models.UserGroup, o *models.User) *orm.Query {
	return m.ForGroupQ(class, group, o).Where("?TableAlias.id = ?", o.ID)
}

// OneByID outputs object filtered by id.
func (m *UserManager) OneByID(o *models.User) error {
	return RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, m.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}

// OneByName outputs object filtered by name.
func (m *UserManager) OneByName(o *models.User) error {
	return RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("u=%s", o.Username), func() (interface{}, error) {
			return o, m.Query(o).Where("username = ?", o.Username).Select()
		}),
	)
}

// OneByKey outputs object filtered by key. Doesn't fetch profile nor groups.
func (m *UserManager) OneByKey(o *models.User) error {
	return RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("k=%s", o.Key), func() (interface{}, error) {
			return o, m.Query(o).Where("key = ?", o.Key).Select()
		}),
	)
}

// FetchData outputs object filtered by name.
func (m *UserManager) FetchData(class *models.Class, o *models.User) error {
	return m.Query(o).Column("_").Relation("Profile").Relation("Groups").
		Where("profile._klass_id = ?", class.ID).WherePK().Select()
}

// CountEstimate returns count estimate for users list.
func (m *UserManager) CountEstimate() (int, error) {
	return CountEstimate(m.dbCtx.Request().Context(), m.DB(), m.Query((*models.User)(nil)), settings.API.DataObjectEstimateThreshold)
}
