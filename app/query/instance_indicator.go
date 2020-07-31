package query

import (
	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// InstanceIndicatorManager represents Instance Indicator manager.
type InstanceIndicatorManager struct {
	*Factory
	*manager.Manager
}

// NewInstanceIndicatorManager creates and returns new Instance Indicator manager.
func (q *Factory) NewInstanceIndicatorManager(c database.DBContext) *InstanceIndicatorManager {
	return &InstanceIndicatorManager{Factory: q, Manager: manager.NewTenantManager(q.db, c)}
}

// ByInstanceAndType filters object filtered by instance and type.
func (m *InstanceIndicatorManager) ByInstanceAndType(o *models.InstanceIndicator) *orm.Query {
	return m.Query(o).Where("instance_id = ?", o.InstanceID).
		Where("type = ?", o.Type)
}
