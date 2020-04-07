package models

import (
	"fmt"
)

// AdminSocialProfile represents AdminSocialProfile model.
type AdminSocialProfile struct {
	tableName struct{} `pg:"admins_adminsocialprofile"` // nolint

	ID       int
	Backend  int
	SocialID int
	AdminID  int
	Admin    *Admin
}

func (m *AdminSocialProfile) String() string {
	return fmt.Sprintf("AdminSocialProfile<ID=%d Backend=%d, Admin=%d>", m.ID, m.Backend, m.AdminID)
}

// VerboseName returns verbose name for model.
func (m *AdminSocialProfile) VerboseName() string {
	return "Admin Social Profile"
}
