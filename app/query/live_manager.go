package query

import (
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/storage"
)

// LiveManager defines a manager with live functionality.
type LiveManager struct {
	*Manager
}

// NewLiveManager creates and returns new live manager.
func NewLiveManager(c storage.DBContext) *LiveManager {
	return &LiveManager{Manager: NewManager(c)}
}

// NewLiveTenantManager creates and returns new live tenant manager.
func NewLiveTenantManager(c storage.DBContext) *LiveManager {
	return &LiveManager{Manager: NewTenantManager(c)}
}

// Query returns only alive objects.
func (m *LiveManager) Query(o interface{}) *orm.Query {
	return m.DB().Model(o).Where("?TableAlias._is_live IS TRUE")
}

// All returns all objects, irrelevant if they are alive or not.
func (m *LiveManager) All() *orm.Query {
	return m.DB().Model()
}

// Dead returns dead objects.
func (m *LiveManager) Dead() *orm.Query {
	return m.DB().Model().Where("?TableAlias._is_live IS NULL")
}

// Delete is a soft delete for live objects.
func (m *LiveManager) Delete(model interface{}) error {
	db := m.DB()
	if _, err := db.Model(model).WherePK().Set("_is_live = ?", false).Update(); err != nil {
		return err
	}

	return storage.ProcessModelSoftDeleteHook(m.Context, db, model)
}

// HardDelete deletes object.
func (m *LiveManager) HardDelete(model interface{}) error {
	return m.Manager.Delete(model)
}
