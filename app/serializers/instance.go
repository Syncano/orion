package serializers

import (
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/fields"
)

type InstanceResponse struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CreatedAt   fields.Time `json:"created_at"`
	UpdatedAt   fields.Time `json:"updated_at"`
	Location    string      `json:"location"`
	Metadata    fields.JSON `json:"metadata"`
	Owner       interface{} `json:"owner"`
}

type InstanceSerializer struct{}

func (s InstanceSerializer) Response(i interface{}) interface{} {
	o := i.(*models.Instance)

	return &InstanceResponse{
		Name:        o.Name,
		Description: o.Description,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
		Location:    o.Location,
		Metadata:    o.Metadata,
		Owner:       AdminSerializer{}.Response(o.Owner),
	}
}
