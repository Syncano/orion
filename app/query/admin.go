package query

import (
	"fmt"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
)

// AdminManager represents Admin manager.
type AdminManager struct {
	*LiveManager
}

// NewAdminManager creates and returns new Admin manager.
func NewAdminManager(c storage.DBContext) *AdminManager {
	return &AdminManager{LiveManager: NewLiveManager(c)}
}

// OneByID outputs object filtered by name.
func (m *AdminManager) OneByID(o *models.Admin) error {
	return RequireOne(
		cache.SimpleModelCache(m.DB(), o, fmt.Sprintf("id=%d", o.ID), func() (interface{}, error) {
			return o, m.Query(o).Where("id = ?", o.ID).Select()
		}),
	)
}

// OneByKey outputs object filtered by key.
func (m *AdminManager) OneByKey(o *models.Admin) error {
	return RequireOne(
		cache.SimpleModelCache(m.DB(), o, fmt.Sprintf("k=%s", o.Key), func() (interface{}, error) {
			return o, m.Query(o).Where("key = ?", o.Key).Select()
		}),
	)
}
