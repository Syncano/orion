package serializers

import (
	"github.com/Syncano/orion/app/models"
)

// UserClassResponse ...
type UserClassResponse struct {
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Schema       *models.JSON `json:"schema"`
	Status       string       `json:"status"`
	CreatedAt    *models.Time `json:"created_at"`
	UpdatedAt    *models.Time `json:"updated_at"`
	ObjectsCount int          `json:"objects_count"`
	Revision     int          `json:"revision"`
	Metadata     *models.JSON `json:"metadata"`
}

// UserClassSerializer ...
type UserClassSerializer struct{}

// Response ...
func (s UserClassSerializer) Response(i interface{}) interface{} {
	o := i.(*models.Class)
	cls := &UserClassResponse{
		Name:         o.Name,
		Description:  o.Description,
		Schema:       &o.Schema,
		Status:       o.GetStatus(),
		CreatedAt:    &o.CreatedAt,
		UpdatedAt:    &o.UpdatedAt,
		ObjectsCount: o.ObjectsCount,
		Revision:     o.Revision,
		Metadata:     &o.Metadata,
	}
	return cls
}
