package query

import (
	"fmt"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
)

// InstanceManager represents Instance manager.
type InstanceManager struct {
	*LiveManager
}

// NewInstanceManager creates and returns new Instance manager.
func NewInstanceManager(c storage.DBContext) *InstanceManager {
	return &InstanceManager{LiveManager: NewLiveManager(c)}
}

// WithAccessQ outputs objects that entity has access to.
func (mgr *InstanceManager) WithAccessQ(o interface{}) *orm.Query {
	q := mgr.Query(o).Column("instance.*", "Owner")
	if a := mgr.Context.Get(settings.ContextAdminKey); a != nil {
		q = q.Join("JOIN admins_admininstancerole AS air ON air.instance_id = instance.id AND air.admin_id = ?", a.(*models.Admin).ID)
	} else if a := mgr.Context.Get(settings.ContextAPIKeyKey); a != nil {
		q = q.Where("id = ?", a.(*models.APIKey).InstanceID)
	}

	return q
}

// WithAccessByNameQ outputs one object that entity has access to filtered by name.
func (mgr *InstanceManager) WithAccessByNameQ(o *models.Instance) *orm.Query {
	return mgr.WithAccessQ(o).
		Where("?TableAlias.name = ?", o.Name)
}

// OneByName outputs object filtered by name.
func (mgr *InstanceManager) OneByName(o *models.Instance) error {
	return RequireOne(
		cache.SimpleModelCache(mgr.DB(), o, fmt.Sprintf("n=%s", o.Name), func() (interface{}, error) {
			return o, mgr.Query(o).
				Where("name = ?", o.Name).Select()
		}),
	)
}
