package query

import (
	"fmt"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/database"
	"github.com/Syncano/pkg-go/database/manager"
)

// AdminInstanceRoleManager represents AdminInstanceRole manager.
type AdminInstanceRoleManager struct {
	*Factory
	*manager.Manager
}

// NewAdminInstanceRoleManager creates and returns new AdminInstanceRole manager.
func (q *Factory) NewAdminInstanceRoleManager(c database.DBContext) *AdminInstanceRoleManager {
	return &AdminInstanceRoleManager{Manager: manager.NewManager(q.db, c)}
}

// OneByInstanceAndAdmin outputs object filtered by instance and admin.
func (m *AdminInstanceRoleManager) OneByInstanceAndAdmin(o *models.AdminInstanceRole) error {
	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("i=%d;a=%d", o.InstanceID, o.AdminID), func() (interface{}, error) {
			return o, m.Query(o).Where("instance_id = ?", o.InstanceID).Where("admin_id = ?", o.AdminID).Select()
		}),
	)
}
