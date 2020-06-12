package serializers

import (
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/database/fields"
)

type UserClassResponse struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Schema       fields.JSON `json:"schema"`
	Status       string      `json:"status"`
	CreatedAt    fields.Time `json:"created_at"`
	UpdatedAt    fields.Time `json:"updated_at"`
	ObjectsCount int         `json:"objects_count"`
	Revision     int         `json:"revision"`
	Metadata     fields.JSON `json:"metadata"`
}

type UserClassSerializer struct{}

func (s UserClassSerializer) Response(i interface{}) interface{} {
	o := i.(*models.Class)
	cls := &UserClassResponse{
		Name:         o.Name,
		Description:  o.Description,
		Schema:       o.Schema,
		Status:       o.GetStatus(),
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
		ObjectsCount: o.ObjectsCount,
		Revision:     o.Revision,
		Metadata:     o.Metadata,
	}

	return cls
}
