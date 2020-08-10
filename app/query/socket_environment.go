package query

import (
	"fmt"

	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// SocketEnvironmentManager represents Socket Environment manager.
type SocketEnvironmentManager struct {
	*Factory
	*manager.LiveManager
}

// NewSocketEnvironmentManager creates and returns new Socket Environment manager.
func (q *Factory) NewSocketEnvironmentManager(c echo.Context) *SocketEnvironmentManager {
	return &SocketEnvironmentManager{Factory: q, LiveManager: manager.NewLiveTenantManager(WrapContext(c), q.db)}
}

// OneByID outputs object filtered by ID.
func (m *SocketEnvironmentManager) OneByID(o *models.SocketEnvironment) error {
	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("i=%d", o.ID), func() (interface{}, error) {
			return o, m.QueryContext(DBToStdContext(m), o).Where("id = ?", o.ID).Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (m *SocketEnvironmentManager) WithAccessQ(o interface{}) *orm.Query {
	return m.QueryContext(DBToStdContext(m), o)
}
