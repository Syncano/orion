package models

import (
	"fmt"

	"github.com/Syncano/pkg-go/database/fields"
)

// Subscription represents Pricing Plan subscription model.
type Subscription struct {
	tableName struct{} `pg:"billing_subscription,discard_unknown_columns"` // nolint

	ID           int
	Commitment   fields.JSON
	ChargedUntil fields.Date
	Range        fields.Daterange

	AdminID int
	Admin   *Admin `msgpack:"-"`
	PlanID  int
	Plan    *PricingPlan
}

func (m *Subscription) String() string {
	return fmt.Sprintf("Subscription<ID=%d Admin=%d Plan=%d>", m.ID, m.AdminID, m.PlanID)
}

// VerboseName returns verbose name for model.
func (m *Subscription) VerboseName() string {
	return "Subscription"
}
