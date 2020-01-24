package query

import (
	"fmt"
	"time"

	"github.com/jinzhu/now"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
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
	*Manager
}

// NewProfileManager creates and returns new Subscription manager.
func NewProfileManager(c storage.DBContext) *ProfileManager {
	return &ProfileManager{Manager: NewManager(c)}
}

// GetBillingStatus returns status string for subscription.
func (mgr *ProfileManager) GetBillingStatus(sub *models.Subscription) (string, error) {
	if sub == nil {
		return BillingStatusNoActiveSubscription, nil
	}

	var ret string

	o := &models.Profile{AdminID: sub.AdminID}
	n := time.Now()
	err := cache.ModelCache(mgr.DB(), o, &ret, fmt.Sprintf("billing;a=%d;t=%s", o.AdminID, n.Format("06-01")), func() (interface{}, error) {
		if b, err := mgr.Query(o).Where("admin_id = ? AND hard_limit_reached >= ?", o.AdminID, now.BeginningOfMonth()).Exists(); err != nil {
			return "", err
		} else if b {
			if !sub.Plan.PaidPlan {
				return BillingStatusFreeLimitExceeded, nil
			}
			return BillingStatusHardLimitExceeded, nil
		}

		if b, err := mgr.Query((*models.Invoice)(nil)).
			Where("admin_id = ? AND due_date <= ? AND status = ?", o.AdminID, n, models.InvoiceStatusPaymentFailed).Exists(); err != nil {
			return "", err
		} else if b {
			return BillingStatusOverdueInvoices, nil
		}
		return "", nil
	}, nil)

	return ret, err
}
