package models

import (
	"fmt"
)

// PricingPlan stores information about prices for given service.
//
// pricing -- For each possible Metric Sources value we define a dictionary with keys from options for this source,
// and dictionary containing plan information.
// Example: {"api": {"20": {"overage": "0.00002", "included":   1000000}, ... }, "cbx: {"5": {"overage": "0.00025", "included":   20000}, ... }}
//
// If there is 'override' key defined in source value it will be used instead of Subscription.commitment value.
// Example: {"api": {"override": {"overage": "0.00002", "included": 200000}}}
//
// options -- For each possible Metric Sources value we define a list of allowed option value.
// Example: {"api": ["20", ... ], "cbx": ["5", ... ]}
type PricingPlan struct {
	tableName struct{} `sql:"billing_pricingplan" pg:",discard_unknown_columns"` // nolint

	ID               int
	Name             string
	Available        bool
	Pricing          *JSON
	Options          *JSON
	AdjustableLimits bool
	PaidPlan         bool
}

func (m *PricingPlan) String() string {
	return fmt.Sprintf("PricingPlan<Name=%q>", m.Name)
}

// VerboseName returns verbose name for model.
func (m *PricingPlan) VerboseName() string {
	return "Pricing Plan"
}
