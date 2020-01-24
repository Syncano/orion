package serializers

import (
	"github.com/Syncano/orion/app/models"
)

type UserGroupResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type UserGroupShortResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
}

type UserGroupSerializer struct{}

func (s UserGroupSerializer) Response(i interface{}) interface{} {
	o := i.(*models.UserGroup)

	return &UserGroupResponse{
		ID:          o.ID,
		Name:        o.Name,
		Label:       o.Label,
		Description: o.Description,
	}
}

func (s UserGroupSerializer) ShortResponse(i interface{}) interface{} {
	o := i.(*models.UserGroup)

	return &UserGroupShortResponse{
		ID:    o.ID,
		Name:  o.Name,
		Label: o.Label,
	}
}
