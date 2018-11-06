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
func (mgr *SubscriptionManager) OneActiveForAdmin(o *models.Subscription, t time.Time) error {
	return cache.ModelCache(mgr.DB(), &models.Profile{AdminID: o.AdminID}, o, fmt.Sprintf("sub;a=%d;t=%s", o.AdminID, t.Format("06-01")), func() (interface{}, error) {
		return o, mgr.Query(o).Column("subscription.*", "Plan").
			Where("admin_id = ? AND range @> ?::date", o.AdminID, t).Select()
	}, nil)
}
