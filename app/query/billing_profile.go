package query

import (
	"fmt"
	"time"

	"github.com/jinzhu/now"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/database"
	"github.com/Syncano/pkg-go/database/manager"
)

// BillingStatus enum.
const (
	BillingStatusNoActiveSubscription = "no_active_subscription"
	BillingStatusHardLimitExceeded    = "hard_limit_reached"
	BillingStatusFreeLimitExceeded    = "free_limits_exceeded"
	BillingStatusOverdueInvoices      = "overdue_invoices"
)

// ProfileManager represents Profile manager.
type ProfileManager struct {
	*Factory
	*manager.Manager
}

// NewProfileManager creates and returns new Subscription manager.
func (q *Factory) NewProfileManager(c database.DBContext) *ProfileManager {
	return &ProfileManager{Factory: q, Manager: manager.NewManager(q.db, c)}
}

// GetBillingStatus returns status string for subscription.
func (m *ProfileManager) GetBillingStatus(sub *models.Subscription) (string, error) {
	if sub == nil {
		return BillingStatusNoActiveSubscription, nil
	}

	var ret string

	o := &models.Profile{AdminID: sub.AdminID}
	n := time.Now()
	err := m.c.ModelCache(m.DB(), o, &ret, fmt.Sprintf("billing;a=%d;t=%s", o.AdminID, n.Format("06-01")), func() (interface{}, error) {
		if b, err := m.Query(o).Where("admin_id = ? AND hard_limit_reached >= ?", o.AdminID, now.BeginningOfMonth()).Exists(); err != nil {
			return "", err
		} else if b {
			if !sub.Plan.PaidPlan {
				return BillingStatusFreeLimitExceeded, nil
			}
			return BillingStatusHardLimitExceeded, nil
		}

		if b, err := m.Query((*models.Invoice)(nil)).
			Where("admin_id = ? AND due_date <= ? AND status = ?", o.AdminID, n, models.InvoiceStatusPaymentFailed).Exists(); err != nil {
			return "", err
		} else if b {
			return BillingStatusOverdueInvoices, nil
		}
		return "", nil
	}, nil)

	return ret, err
}
