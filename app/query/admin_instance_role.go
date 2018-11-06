package query

import (
	"fmt"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
)

// AdminInstanceRoleManager represents AdminInstanceRole manager.
type AdminInstanceRoleManager struct {
	*Manager
}

// NewAdminInstanceRoleManager creates and returns new AdminInstanceRole manager.
func NewAdminInstanceRoleManager(c storage.DBContext) *AdminInstanceRoleManager {
	return &AdminInstanceRoleManager{Manager: NewManager(c)}
}

// OneByInstanceAndAdmin outputs object filtered by instance and admin.
func (mgr *AdminInstanceRoleManager) OneByInstanceAndAdmin(o *models.AdminInstanceRole) error {
	return RequireOne(
		cache.SimpleModelCache(mgr.DB(), o, fmt.Sprintf("i=%d;a=%d", o.InstanceID, o.AdminID), func() (interface{}, error) {
			return o, mgr.Query(o).Where("instance_id = ?", o.InstanceID).Where("admin_id = ?", o.AdminID).Select()
		}),
	)
}
