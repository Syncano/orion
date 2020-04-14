package query

import (
	"fmt"
	"time"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
)

// SubscriptionManager represents Subscription manager.
type SubscriptionManager struct {
	*Manager
}

// NewSubscriptionManager creates and returns new Subscription manager.
func NewSubscriptionManager(c storage.DBContext) *SubscriptionManager {
	return &SubscriptionManager{Manager: NewManager(c)}
}

// OneActiveForAdmin returns subscription active at time for specified o.AdminID.
func (m *SubscriptionManager) OneActiveForAdmin(o *models.Subscription, t time.Time) error {
	return cache.ModelCache(m.DB(), &models.Profile{AdminID: o.AdminID}, o, fmt.Sprintf("sub;a=%d;t=%s", o.AdminID, t.Format("06-01")), func() (interface{}, error) {
		return o, m.Query(o).Column("subscription.*").Relation("Plan").
			Where("admin_id = ? AND range @> ?::date", o.AdminID, t).Select()
	}, nil)
}
