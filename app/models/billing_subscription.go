package models

import (
	"fmt"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/cache"
)

// Subscription represents Pricing Plan subscription model.
type Subscription struct {
	tableName struct{} `sql:"billing_subscription" pg:",discard_unknown_columns"` // nolint

	ID           int
	Commitment   JSON
	ChargedUntil Date
	Range        Daterange

	AdminID int
	Admin   *Admin `msgpack:"-"`
	PlanID  int
	Plan    *PricingPlan `msgpack:"-"`
}

func (m *Subscription) String() string {
	return fmt.Sprintf("Subscription<ID=%d Admin=%d Plan=%d>", m.ID, m.AdminID, m.PlanID)
}

// VerboseName returns verbose name for model.
func (m *Subscription) VerboseName() string {
	return "Subscription"
}

// AfterUpdate hook.
func (m *Subscription) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *Subscription) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}
