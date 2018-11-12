package query

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
)

// SocketEndpointManager represents Socket Endpoint manager.
type SocketEndpointManager struct {
	*Manager
}

// NewSocketEndpointManager creates and returns new Socket Endpoint manager.
func NewSocketEndpointManager(c storage.DBContext) *SocketEndpointManager {
	return &SocketEndpointManager{Manager: NewTenantManager(c)}
}

// ForSocketQ outputs object filtered by name.
func (mgr *SocketEndpointManager) ForSocketQ(socket *models.Socket, o interface{}) *orm.Query {
	return mgr.Query(o).Where("socket_id = ?", socket.ID)
}

// OneByName outputs object filtered by name.
func (mgr *SocketEndpointManager) OneByName(o *models.SocketEndpoint) error {
	o.Name = strings.ToLower(o.Name)
	return RequireOne(
		cache.SimpleModelCache(mgr.DB(), o, fmt.Sprintf("n=%s", o.Name), func() (interface{}, error) {
			return o, mgr.Query(o).Where("name = ?", o.Name).Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (mgr *SocketEndpointManager) WithAccessQ(o interface{}) *orm.Query {
	return mgr.Query(o)
}
