package models

import (
	"fmt"

	"github.com/go-pg/pg/orm"
	"github.com/shopspring/decimal"

	"github.com/Syncano/orion/pkg/cache"
)

// Profile represents billing profile model.
type Profile struct {
	State
	tableName struct{} `sql:"billing_profile" pg:",discard_unknown_columns"` // nolint

	AdminID          int `sql:",pk"`
	Admin            *Admin
	CustomerID       string
	SoftLimit        decimal.Decimal
	SoftLimitReached *Date
	HardLimit        decimal.Decimal
	HardLimitReached *Date

	CompanyName    string
	FirstName      string
	LastName       string
	TaxNumber      string
	AddressCity    string
	AddressCountry string
	AddressLine1   string
	AddressLine2   string
	AddressState   string
	AddressZip     string
}

func (m *Profile) String() string {
	return fmt.Sprintf("Profile<Admin=%d>", m.AdminID)
}

// VerboseName returns verbose name for model.
func (m *Profile) VerboseName() string {
	return "Profile"
}

// AfterUpdate hook.
func (m *Profile) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *Profile) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}
