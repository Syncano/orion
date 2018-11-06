package serializers

import (
	"github.com/Syncano/orion/app/models"
)

// InstanceResponse ...
type InstanceResponse struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	CreatedAt   *models.Time `json:"created_at"`
	UpdatedAt   *models.Time `json:"updated_at"`
	Location    string       `json:"location"`
	Metadata    *models.JSON `json:"metadata"`
	Owner       interface{}  `json:"owner"`
}

// InstanceSerializer ...
type InstanceSerializer struct{}

// Response ...
func (s InstanceSerializer) Response(i interface{}) interface{} {
	o := i.(*models.Instance)
	return &InstanceResponse{
		Name:        o.Name,
		Description: o.Description,
		CreatedAt:   &o.CreatedAt,
		UpdatedAt:   &o.UpdatedAt,
		Location:    o.Location,
		Metadata:    &o.Metadata,
		Owner:       AdminSerializer{}.Response(o.Owner),
	}
}
