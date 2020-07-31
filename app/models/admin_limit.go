package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgtype"

	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/v2/database/fields"
)

const freePlanName = "free"

// AdminLimit represents admin limit model.
type AdminLimit struct {
	tableName struct{} `pg:"billing_adminlimit,discard_unknown_columns"` // nolint

	AdminID int `pg:",pk"`
	Admin   *Admin
	Limits  fields.Hstore
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

func (m *AdminLimit) RateLimit(sub *Subscription) *settings.RateData {
	return &settings.RateData{
		Limit:    int64(m.getLimit(sub, "rate", settings.Billing.RateLimit)),
		Duration: time.Second,
	}
}

func (m *AdminLimit) PollRateLimit(sub *Subscription) *settings.RateData {
	return &settings.RateData{
		Limit:    int64(m.getLimit(sub, "poll_rate", settings.Billing.PollRateLimit)),
		Duration: time.Second,
	}
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
