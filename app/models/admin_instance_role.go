package models

import (
	"fmt"
)

// AdminInstanceRole represents AdminInstanceRole model.
type AdminInstanceRole struct {
	tableName struct{} `pg:"admins_admininstancerole"` // nolint

	ID         int
	AdminID    int
	Admin      *Admin
	InstanceID int
	Instance   *Instance
}

func (m *AdminInstanceRole) String() string {
	return fmt.Sprintf("AdminInstanceRole<Admin=%d Instance=%d>", m.AdminID, m.InstanceID)
}

// VerboseName returns verbose name for model.
func (m *AdminInstanceRole) VerboseName() string {
	return "Admin Instance Role"
}
