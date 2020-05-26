package query

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/storage"
)

// SocketManager represents Socket manager.
type SocketManager struct {
	*LiveManager
}

// NewSocketManager creates and returns new Socket manager.
func (q *Factory) NewSocketManager(c storage.DBContext) *SocketManager {
	return &SocketManager{LiveManager: q.NewLiveTenantManager(c)}
}

// OneByID outputs object filtered by ID.
func (m *SocketManager) OneByID(o *models.Socket) error {
	return RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, m.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}

// OneByName outputs object filtered by name.
func (m *SocketManager) OneByName(o *models.Socket) error {
	o.Name = strings.ToLower(o.Name)

	return RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("n=%s", o.Name), func() (interface{}, error) {
			return o, m.Query(o).Where("name = ?", o.Name).Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (m *SocketManager) WithAccessQ(o interface{}) *orm.Query {
	return m.Query(o)
}
