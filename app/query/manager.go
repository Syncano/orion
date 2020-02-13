package query

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/storage"
)

// Manager defines object manager.
type Manager struct {
	Context storage.DBContext
	db      orm.DB
	dbGet   func(storage.DBContext) orm.DB
}

// NewManager creates and returns new manager.
func NewManager(c storage.DBContext) *Manager {
	return &Manager{Context: c, dbGet: func(c storage.DBContext) orm.DB {
		return DB(c)
	}}
}

// NewTenantManager creates and returns new tenant manager.
func NewTenantManager(c storage.DBContext) *Manager {
	return &Manager{Context: c, dbGet: func(c storage.DBContext) orm.DB {
		return TenantDB(c)
	}}
}

// DB returns standard database.
func (m *Manager) DB() orm.DB {
	if m.db != nil {
		return m.db
	}

	return m.dbGet(m.Context)
}

// SetDB sets database.
func (m *Manager) SetDB(db orm.DB) {
	m.db = db
}

// Query returns all objects.
func (m *Manager) Query(o interface{}) *orm.Query {
	return m.DB().Model(o)
}

// Insert creates object.
func (m *Manager) Insert(model interface{}) error {
	db := m.DB()
	if err := db.Insert(model); err != nil {
		return err
	}

	return storage.ProcessModelSaveHook(m.Context, db, true, model)
}

// Update updates object.
func (m *Manager) Update(model interface{}, fields ...string) error {
	db := m.DB()
	if _, err := db.Model(model).Column(fields...).WherePK().Update(); err != nil {
		return err
	}

	return storage.ProcessModelSaveHook(m.Context, db, false, model)
}

// Delete deletes object.
func (m *Manager) Delete(model interface{}) error {
	db := m.DB()
	if err := db.Delete(model); err != nil {
		return err
	}

	return storage.ProcessModelDeleteHook(m.Context, db, model)
}

// RunInTransaction is an alias for DB function.
func (m *Manager) RunInTransaction(fn func(*pg.Tx) error) error {
	var (
		tx  *pg.Tx
		err error
	)

	if m.db == nil {
		tx, err = m.dbGet(m.Context).(*pg.DB).Begin()
		if err != nil {
			return err
		}

		m.db = tx

		defer func() {
			m.db = nil
		}()
	}

	return storage.RunTransactionWithHooks(tx, fn)
}
