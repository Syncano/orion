package query

import (
	"fmt"
	"time"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

// SubscriptionManager represents Subscription manager.
type SubscriptionManager struct {
	*Factory
	*manager.Manager
}

// NewSubscriptionManager creates and returns new Subscription manager.
func (q *Factory) NewSubscriptionManager(c database.DBContext) *SubscriptionManager {
	return &SubscriptionManager{Factory: q, Manager: manager.NewManager(q.db, c)}
}

// OneActiveForAdmin returns subscription active at time for specified o.AdminID.
func (m *SubscriptionManager) OneActiveForAdmin(o *models.Subscription, t time.Time) error {
	return m.c.ModelCache(m.DB(), &models.Profile{AdminID: o.AdminID}, o, fmt.Sprintf("sub;a=%d;t=%s", o.AdminID, t.Format("06-01")), func() (interface{}, error) {
		return o, m.Query(o).Column("subscription.*").Relation("Plan").
			Where("admin_id = ? AND range @> ?::date", o.AdminID, t).Select()
	}, nil)
}
