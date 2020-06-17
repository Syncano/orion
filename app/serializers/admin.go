package serializers

import (
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/database/fields"
)

type AdminResponse struct {
	ID          int         `json:"id"`
	Email       string      `json:"email"`
	FirstName   string      `json:"first_name"`
	LastName    string      `json:"last_name"`
	IsActive    bool        `json:"is_active"`
	HasPassword bool        `json:"has_password"`
	Metadata    fields.JSON `json:"metadata"`
}

type AdminSerializer struct{}

func (s AdminSerializer) Response(i interface{}) interface{} {
	o := i.(*models.Admin)

	return &AdminResponse{
		ID:          o.ID,
		Email:       o.Email,
		FirstName:   o.FirstName,
		LastName:    o.LastName,
		IsActive:    o.IsActive,
		HasPassword: o.IsPasswordUsable(),
		Metadata:    o.Metadata,
	}
}
