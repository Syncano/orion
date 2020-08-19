package query

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// AdminManager represents Admin manager.
type AdminManager struct {
	*Factory
	*manager.LiveManager
}

// NewAdminManager creates and returns new Admin manager.
func (q *Factory) NewAdminManager(c echo.Context) *AdminManager {
	return &AdminManager{Factory: q, LiveManager: manager.NewLiveManager(WrapContext(c), q.db)}
}

// OneByID outputs object filtered by name.
func (m *AdminManager) OneByID(o *models.Admin) error {
	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("id=%d", o.ID), func() (interface{}, error) {
			return o, m.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}

// OneByKey outputs object filtered by key.
func (m *AdminManager) OneByKey(o *models.Admin) error {
	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("k=%s", o.Key), func() (interface{}, error) {
			return o, m.Query(o).Where("key = ?", o.Key).Select()
		}),
	)
}
