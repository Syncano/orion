package models

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/Syncano/pkg-go/v2/database/fields"
)

// Profile represents billing profile model.
type Profile struct {
	State
	tableName struct{} `pg:"billing_profile,discard_unknown_columns"` // nolint

	AdminID          int `pg:",pk"`
	Admin            *Admin
	CustomerID       string
	SoftLimit        decimal.Decimal
	SoftLimitReached fields.Date
	HardLimit        decimal.Decimal
	HardLimitReached fields.Date

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
