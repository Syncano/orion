package query

import (
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// DataObjectManager represents DataObject manager.
type DataObjectManager struct {
	*Factory
	*manager.LiveManager
}

// NewDataObjectManager creates and returns new DataObject manager.
func (q *Factory) NewDataObjectManager(c echo.Context) *DataObjectManager {
	return &DataObjectManager{Factory: q, LiveManager: manager.NewLiveTenantManager(WrapContext(c), q.db)}
}

// Create creates new object.
func (m *DataObjectManager) Create(o interface{}) error {
	_, e := m.DB().ModelContext(DBToStdContext(m), o).Returning("*").Insert()
	return e
}

// CountEstimate returns count estimate for current data objects list.
func (m *DataObjectManager) CountEstimate(q orm.QueryAppender) (int, error) {
	return manager.CountEstimate(DBToStdContext(m), m.DB(), q, settings.API.DataObjectEstimateThreshold)
}

// ForClassQ outputs objects within specific class.
func (m *DataObjectManager) ForClassQ(class *models.Class, o interface{}) *orm.Query {
	q := m.QueryContext(DBToStdContext(m), o).Where("_klass_id = ?", class.ID)
	return q
}

// ForClassByIDQ outputs one object within specific class filtered by id.
func (m *DataObjectManager) ForClassByIDQ(class *models.Class, o *models.DataObject) *orm.Query {
	return m.ForClassQ(class, o).Where("?TableAlias.id = ?", o.ID)
}
