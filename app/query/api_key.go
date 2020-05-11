package query

import (
	"fmt"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/storage"
)

// APIKeyManager represents APIKey manager.
type APIKeyManager struct {
	*LiveManager
}

// NewAPIKeyManager creates and returns new APIKey manager.
func (q *Factory) NewAPIKeyManager(c storage.DBContext) *APIKeyManager {
	return &APIKeyManager{LiveManager: q.NewLiveManager(c)}
}

// OneByKey outputs object filtered by key.
func (m *APIKeyManager) OneByKey(o *models.APIKey) error {
	return RequireOne(
		m.c.SimpleModelCache(m.DB(), o, fmt.Sprintf("k=%s", o.Key), func() (interface{}, error) {
			return o, m.Query(o).Where("key = ?", o.Key).Select()
		}),
	)
}
