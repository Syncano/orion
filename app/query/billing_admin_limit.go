package query

import (
	"fmt"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
)

// AdminLimitManager represents Admin Limit manager.
type AdminLimitManager struct {
	*Manager
}

// NewAdminLimitManager creates and returns new Admin Limit manager.
func NewAdminLimitManager(c storage.DBContext) *AdminLimitManager {
	return &AdminLimitManager{Manager: NewManager(c)}
}

// OneForAdmin returns admin limit for specified o.AdminID.
func (m *AdminLimitManager) OneForAdmin(o *models.AdminLimit) error {
	return cache.SimpleModelCache(m.DB(), o, fmt.Sprintf("a=%d", o.AdminID), func() (interface{}, error) {
		return o, m.Query(o).Where("admin_id = ?", o.AdminID).Select()
	})
}
