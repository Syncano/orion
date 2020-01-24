package models

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx/pgtype"

	"github.com/Syncano/orion/pkg/settings"
)

const freePlanName = "free"

// AdminLimit represents admin limit model.
type AdminLimit struct {
	tableName struct{} `sql:"billing_adminlimit" pg:",discard_unknown_columns"` // nolint

	AdminID int `sql:",pk"`
	Admin   *Admin
	Limits  Hstore
}

func (m *AdminLimit) String() string {
	return fmt.Sprintf("AdminLimit<Admin=%d>", m.AdminID)
}

// VerboseName returns verbose name for model.
func (m *AdminLimit) VerboseName() string {
	return "Admin Limit"
}

func (m *AdminLimit) getLimit(sub *Subscription, key string, limit settings.PlanLimit) int {
	if !m.Limits.IsNull() {
		if v, ok := m.Limits.Map[key]; ok && v.Status == pgtype.Present {
			if i, err := strconv.Atoi(v.String); err == nil {
				return i
			}
		}
	}

	if sub == nil || sub.Plan == nil {
		return limit.Default
	}

	if sub.Plan.PaidPlan || sub.Plan.Name == freePlanName {
		return limit.Paid
	}

	return limit.Builder
}

func (m *AdminLimit) StorageLimit(sub *Subscription) int {
	return m.getLimit(sub, "storage", settings.Billing.StorageLimit)
}

func (m *AdminLimit) RateLimit(sub *Subscription) int {
	return m.getLimit(sub, "rate", settings.Billing.RateLimit)
}

func (m *AdminLimit) PollRateLimit(sub *Subscription) int {
	return m.getLimit(sub, "poll_rate", settings.Billing.PollRateLimit)
}

func (m *AdminLimit) CodeboxConcurrency(sub *Subscription) int {
	return m.getLimit(sub, "codebox_concurrency", settings.Billing.CodeboxConcurrency)
}

func (m *AdminLimit) ClassesCount(sub *Subscription) int {
	return m.getLimit(sub, "classes_count", settings.Billing.ClassesCount)
}

func (m *AdminLimit) SocketsCount(sub *Subscription) int {
	return m.getLimit(sub, "sockets_count", settings.Billing.SocketsCount)
}

func (m *AdminLimit) SchedulesCount(sub *Subscription) int {
	return m.getLimit(sub, "schedules_count", settings.Billing.SchedulesCount)
}

func (m *AdminLimit) InstancesCount(sub *Subscription) int {
	return m.getLimit(sub, "instances_count", settings.Billing.InstancesCount)
}
