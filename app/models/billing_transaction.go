package models

import (
	"fmt"
)

// Transaction represents billing transaction model.
type Transaction struct {
	tableName struct{} `sql:"billing_transaction" pg:",discard_unknown_columns"` // nolint

	ID           int
	AdminID      int
	Admin        *Admin
	InstanceID   int
	InstanceName string
	Source       string
	Quantity     int
	Period       *Time
	Aggregated   bool
	CreatedAt    *Time
}

func (m *Transaction) String() string {
	return fmt.Sprintf("Transaction<ID=%d Admin=%d Period=%q>", m.ID, m.AdminID, m.Period.String())
}

// VerboseName returns verbose name for model.
func (m *Transaction) VerboseName() string {
	return "Transaction"
}
