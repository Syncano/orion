package models

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/Syncano/pkg-go/database/fields"
)

// InvoiceStatus enum.
const (
	InvoiceStatusNew int = iota
	InvoiceStatusPending
	InvoiceStatusFake
	InvoiceStatusEmpty
	InvoiceStatusSchedulingFailed
	InvoiceStatusPaymentScheduled
	InvoiceStatusPaymentFailed
	InvoiceStatusPaymentSucceeded
)

// InvoiceStatus to string map.
var InvoiceStatus = map[int]string{
	InvoiceStatusNew:              "new",
	InvoiceStatusPending:          "pending",
	InvoiceStatusFake:             "fake",
	InvoiceStatusEmpty:            "empty",
	InvoiceStatusSchedulingFailed: "scheduling failed",
	InvoiceStatusPaymentScheduled: "payment scheduled",
	InvoiceStatusPaymentFailed:    "payment failed",
	InvoiceStatusPaymentSucceeded: "payment succeeded",
}

// Invoice represents billing invoice model.
type Invoice struct {
	tableName struct{} `pg:"billing_invoice,discard_unknown_columns"` // nolint

	ID            int
	AdminID       int
	Admin         *Admin
	Status        int
	PlanFee       decimal.Decimal
	OverageAmount decimal.Decimal
	Period        fields.Date
	IsProrated    bool
	DueDate       fields.Date
	ExternalID    string
	CreatedAt     fields.Time
	UpdatedAt     fields.Time
	Reference     string
	StatusSent    bool
}

func (m *Invoice) String() string {
	return fmt.Sprintf("Invoice<ID=%d, Admin=%d, Status=%q>", m.ID, m.AdminID, m.StatusString())
}

// VerboseName returns verbose name for model.
func (m *Invoice) VerboseName() string {
	return "Invoice"
}

// StatusString returns verbose value for status.
func (m *Invoice) StatusString() string {
	return InvoiceStatus[m.Status]
}

// BeforeUpdate hook.
func (m *Invoice) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return ctx, nil
}

// InvoiceItemSource enum.
const (
	InvoiceItemSourcePlanFee = "fee"
	InvoiceItemSourceAPI     = "api"
	InvoiceItemSourceCodebox = "cbx"
)

// InvoiceItemSource to string map.
var InvoiceItemSource = map[string]string{
	InvoiceItemSourcePlanFee: "Plan Fee",
	InvoiceItemSourceAPI:     "API Call",
	InvoiceItemSourceCodebox: "Script Execution Time (s)",
}

// InvoiceItem represents billing invoice item model.
type InvoiceItem struct {
	tableName struct{} `pg:"billing_invoiceitem,discard_unknown_columns"` // nolint

	ID           int
	InvoiceID    int
	Invoice      *Invoice
	InstanceName string
	InstanceID   int

	Source    string
	Quantity  int
	Price     decimal.Decimal
	CreatedAt fields.Time
	UpdatedAt fields.Time
}

func (m *InvoiceItem) String() string {
	return fmt.Sprintf("InvoiceItem<ID=%d, Invoice=%d>", m.ID, m.InvoiceID)
}

// VerboseName returns verbose name for model.
func (m *InvoiceItem) VerboseName() string {
	return "Invoice Item"
}

// SourceString returns verbose value for source.
func (m *InvoiceItem) SourceString() string {
	return InvoiceItemSource[m.Source]
}
