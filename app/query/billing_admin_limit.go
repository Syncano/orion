package query

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// AdminLimitManager represents Admin Limit manager.
type AdminLimitManager struct {
	*Factory
	*manager.Manager
}

// NewAdminLimitManager creates and returns new Admin Limit manager.
func (q *Factory) NewAdminLimitManager(c echo.Context) *AdminLimitManager {
	return &AdminLimitManager{Factory: q, Manager: manager.NewManager(WrapContext(c), q.db)}
}

// OneForAdmin returns admin limit for specified o.AdminID.
func (m *AdminLimitManager) OneForAdmin(o *models.AdminLimit) error {
	return m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("a=%d", o.AdminID), func() (interface{}, error) {
		return o, m.QueryContext(DBToStdContext(m), o).Where("admin_id = ?", o.AdminID).Select()
	})
}
