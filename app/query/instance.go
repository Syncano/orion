package query

import (
	"fmt"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/database"
	"github.com/Syncano/pkg-go/database/manager"
)

// InstanceManager represents Instance manager.
type InstanceManager struct {
	*Factory
	*manager.LiveManager
}

// NewInstanceManager creates and returns new Instance manager.
func (q *Factory) NewInstanceManager(c database.DBContext) *InstanceManager {
	return &InstanceManager{Factory: q, LiveManager: manager.NewLiveManager(q.db, c)}
}

// WithAccessQ outputs objects that entity has access to.
func (m *InstanceManager) WithAccessQ(o interface{}) *orm.Query {
	q := m.Query(o).Column("instance.*").Relation("Owner")
	if a := m.DBContext().Get(settings.ContextAdminKey); a != nil {
		q = q.Join("JOIN admins_admininstancerole AS air ON air.instance_id = instance.id AND air.admin_id = ?", a.(*models.Admin).ID)
	} else if a := m.DBContext().Get(settings.ContextAPIKeyKey); a != nil {
		q = q.Where("id = ?", a.(*models.APIKey).InstanceID)
	}

	return q
}

// WithAccessByNameQ outputs one object that entity has access to filtered by name.
func (m *InstanceManager) WithAccessByNameQ(o *models.Instance) *orm.Query {
	return m.WithAccessQ(o).
		Where("?TableAlias.name = ?", o.Name)
}

// OneByName outputs object filtered by name.
func (m *InstanceManager) OneByName(o *models.Instance) error {
	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("n=%s", o.Name), func() (interface{}, error) {
			return o, m.Query(o).
				Where("name = ?", o.Name).Select()
		}),
	)
}
