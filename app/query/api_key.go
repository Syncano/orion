package query

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// APIKeyManager represents APIKey manager.
type APIKeyManager struct {
	*Factory
	*manager.LiveManager
}

// NewAPIKeyManager creates and returns new APIKey manager.
func (q *Factory) NewAPIKeyManager(c echo.Context) *APIKeyManager {
	return &APIKeyManager{Factory: q, LiveManager: manager.NewLiveManager(WrapContext(c), q.db)}
}

// OneByKey outputs object filtered by key.
func (m *APIKeyManager) OneByKey(o *models.APIKey) error {
	return manager.RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("k=%s", o.Key), func() (interface{}, error) {
			return o, m.Query(o).Where("key = ?", o.Key).Select()
		}),
	)
}
