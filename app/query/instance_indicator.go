package query

import (
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// InstanceIndicatorManager represents Instance Indicator manager.
type InstanceIndicatorManager struct {
	*Factory
	*manager.Manager
}

// NewInstanceIndicatorManager creates and returns new Instance Indicator manager.
func (q *Factory) NewInstanceIndicatorManager(c echo.Context) *InstanceIndicatorManager {
	return &InstanceIndicatorManager{Factory: q, Manager: manager.NewTenantManager(WrapContext(c), q.db)}
}

// ByInstanceAndTypeQ filters object filtered by instance and type.
func (m *InstanceIndicatorManager) ByInstanceAndTypeQ(o *models.InstanceIndicator) *orm.Query {
	return m.QueryContext(DBToStdContext(m), o).Where("instance_id = ?", o.InstanceID).
		Where("type = ?", o.Type)
}
