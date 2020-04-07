package query

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
)

// ChannelManager represents Channel manager.
type ChannelManager struct {
	*LiveManager
}

// NewChannelManager creates and returns new Channel manager.
func NewChannelManager(c storage.DBContext) *ChannelManager {
	return &ChannelManager{LiveManager: NewLiveTenantManager(c)}
}

// OneByName outputs object filtered by name.
func (m *ChannelManager) OneByName(o *models.Channel) error {
	o.Name = strings.ToLower(o.Name)

	return RequireOne(
		cache.SimpleModelCache(m.DB(), o, fmt.Sprintf("n=%s", o.Name), func() (interface{}, error) {
			return o, m.Query(o).Where("name = ?", o.Name).Select()
		}),
	)
}

// WithAccessQ outputs objects that entity has access to.
func (m *ChannelManager) WithAccessQ(o interface{}) *orm.Query {
	return m.Query(o)
}
