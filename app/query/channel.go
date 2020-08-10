package query

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// ChannelManager represents Channel manager.
type ChannelManager struct {
	*Factory
	*manager.LiveManager
}

// NewChannelManager creates and returns new Channel manager.
func (q *Factory) NewChannelManager(c echo.Context) *ChannelManager {
	return &ChannelManager{Factory: q, LiveManager: manager.NewLiveTenantManager(WrapContext(c), q.db)}
}

// OneByName outputs object filtered by name.
func (m *ChannelManager) OneByName(o *models.Channel) error {
	o.Name = strings.ToLower(o.Name)

	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("n=%s", o.Name), func() (interface{}, error) {
			return o, m.Query(o).Where("name = ?", o.Name).Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (m *ChannelManager) WithAccessQ(o interface{}) *orm.Query {
	return m.Query(o)
}
