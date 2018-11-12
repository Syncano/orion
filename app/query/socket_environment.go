package query

import (
	"fmt"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
)

// SocketEnvironmentManager represents Socket Environment manager.
type SocketEnvironmentManager struct {
	*LiveManager
}

// NewSocketEnvironmentManager creates and returns new Socket Environment manager.
func NewSocketEnvironmentManager(c storage.DBContext) *SocketEnvironmentManager {
	return &SocketEnvironmentManager{LiveManager: NewLiveTenantManager(c)}
}

// OneByID outputs object filtered by ID.
func (mgr *SocketEnvironmentManager) OneByID(o *models.SocketEnvironment) error {
	return RequireOne(
		cache.SimpleModelCache(mgr.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, mgr.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (mgr *SocketEnvironmentManager) WithAccessQ(o interface{}) *orm.Query {
	return mgr.Query(o)
}
