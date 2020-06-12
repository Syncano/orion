package query

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/database"
	"github.com/Syncano/pkg-go/database/manager"
)

// SocketEndpointManager represents Socket Endpoint manager.
type SocketEndpointManager struct {
	*Factory
	*manager.Manager
}

// NewSocketEndpointManager creates and returns new Socket Endpoint manager.
func (q *Factory) NewSocketEndpointManager(c database.DBContext) *SocketEndpointManager {
	return &SocketEndpointManager{Manager: manager.NewTenantManager(q.db, c)}
}

// ForSocketQ outputs object filtered by name.
func (m *SocketEndpointManager) ForSocketQ(socket *models.Socket, o interface{}) *orm.Query {
	return m.Query(o).Where("socket_id = ?", socket.ID)
}

// OneByName outputs object filtered by name.
func (m *SocketEndpointManager) OneByName(o *models.SocketEndpoint) error {
	o.Name = strings.ToLower(o.Name)

	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("n=%s", o.Name), func() (interface{}, error) {
			return o, m.Query(o).Where("name = ?", o.Name).Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (m *SocketEndpointManager) WithAccessQ(o interface{}) *orm.Query {
	return m.Query(o)
}
