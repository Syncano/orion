package query

import (
	"fmt"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/storage"
)

// SocketEnvironmentManager represents Socket Environment manager.
type SocketEnvironmentManager struct {
	*LiveManager
}

// NewSocketEnvironmentManager creates and returns new Socket Environment manager.
func (q *Factory) NewSocketEnvironmentManager(c storage.DBContext) *SocketEnvironmentManager {
	return &SocketEnvironmentManager{LiveManager: q.NewLiveTenantManager(c)}
}

// OneByID outputs object filtered by ID.
func (m *SocketEnvironmentManager) OneByID(o *models.SocketEnvironment) error {
	return RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, m.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (m *SocketEnvironmentManager) WithAccessQ(o interface{}) *orm.Query {
	return m.Query(o)
}
