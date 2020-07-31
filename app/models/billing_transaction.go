package models

import (
	"fmt"

	"github.com/Syncano/pkg-go/v2/database/fields"
)

// Transaction represents billing transaction model.
type Transaction struct {
	tableName struct{} `pg:"billing_transaction,discard_unknown_columns"` // nolint

	ID           int
	AdminID      int
	Admin        *Admin
	InstanceID   int
	InstanceName string
	Source       string
	Quantity     int
	Period       fields.Time
	Aggregated   bool
	CreatedAt    fields.Time
}

func (m *Transaction) String() string {
	return fmt.Sprintf("Transaction<ID=%d Admin=%d Period=%q>", m.ID, m.AdminID, m.Period.String())
}

// VerboseName returns verbose name for model.
func (m *Transaction) VerboseName() string {
	return "Transaction"
}
