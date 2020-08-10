package query

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// AdminInstanceRoleManager represents AdminInstanceRole manager.
type AdminInstanceRoleManager struct {
	*Factory
	*manager.Manager
}

// NewAdminInstanceRoleManager creates and returns new AdminInstanceRole manager.
func (q *Factory) NewAdminInstanceRoleManager(c echo.Context) *AdminInstanceRoleManager {
	return &AdminInstanceRoleManager{Factory: q, Manager: manager.NewManager(WrapContext(c), q.db)}
}

// OneByInstanceAndAdmin outputs object filtered by instance and admin.
func (m *AdminInstanceRoleManager) OneByInstanceAndAdmin(o *models.AdminInstanceRole) error {
	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("i=%d;a=%d", o.InstanceID, o.AdminID), func() (interface{}, error) {
			return o, m.QueryContext(DBToStdContext(m), o).Where("instance_id = ?", o.InstanceID).Where("admin_id = ?", o.AdminID).Select()
		}),
	)
}
