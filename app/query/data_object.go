package query

import (
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
)

// DataObjectManager represents DataObject manager.
type DataObjectManager struct {
	*LiveManager
}

// NewDataObjectManager creates and returns new DataObject manager.
func NewDataObjectManager(c storage.DBContext) *DataObjectManager {
	return &DataObjectManager{LiveManager: NewLiveTenantManager(c)}
}

// Create creates new object.
func (mgr *DataObjectManager) Create(o interface{}) error {
	_, e := mgr.DB().Model(o).Returning("*").Insert()
	return e
}

// CountEstimate returns count estimate for current data objects list.
func (mgr *DataObjectManager) CountEstimate(q *orm.Query) (int, error) {
	return CountEstimate(mgr.DB(), q, settings.API.DataObjectEstimateThreshold)
}

// ForClassQ outputs objects within specific class.
func (mgr *DataObjectManager) ForClassQ(class *models.Class, o interface{}) *orm.Query {
	q := mgr.Query(o).Where("_klass_id = ?", class.ID)
	return q
}

// ForClassByIDQ outputs one object within specific class filtered by id.
func (mgr *DataObjectManager) ForClassByIDQ(class *models.Class, o *models.DataObject) *orm.Query {
	return mgr.ForClassQ(class, o).Where("?TableAlias.id = ?", o.ID)
}
