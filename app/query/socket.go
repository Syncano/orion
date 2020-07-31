package query

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// SocketManager represents Socket manager.
type SocketManager struct {
	*Factory
	*manager.LiveManager
}

// NewSocketManager creates and returns new Socket manager.
func (q *Factory) NewSocketManager(c database.DBContext) *SocketManager {
	return &SocketManager{Factory: q, LiveManager: manager.NewLiveTenantManager(q.db, c)}
}

// OneByID outputs object filtered by ID.
func (m *SocketManager) OneByID(o *models.Socket) error {
	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, m.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}

// OneByName outputs object filtered by name.
func (m *SocketManager) OneByName(o *models.Socket) error {
	o.Name = strings.ToLower(o.Name)

	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("n=%s", o.Name), func() (interface{}, error) {
			return o, m.Query(o).Where("name = ?", o.Name).Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (m *SocketManager) WithAccessQ(o interface{}) *orm.Query {
	return m.Query(o)
}
