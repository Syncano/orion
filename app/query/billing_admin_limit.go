package query

import (
	"fmt"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/database"
	"github.com/Syncano/pkg-go/database/manager"
)

// AdminLimitManager represents Admin Limit manager.
type AdminLimitManager struct {
	*Factory
	*manager.Manager
}

// NewAdminLimitManager creates and returns new Admin Limit manager.
func (q *Factory) NewAdminLimitManager(c database.DBContext) *AdminLimitManager {
	return &AdminLimitManager{Manager: manager.NewManager(q.db, c)}
}

// OneForAdmin returns admin limit for specified o.AdminID.
func (m *AdminLimitManager) OneForAdmin(o *models.AdminLimit) error {
	return m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("a=%d", o.AdminID), func() (interface{}, error) {
		return o, m.Query(o).Where("admin_id = ?", o.AdminID).Select()
	})
}
