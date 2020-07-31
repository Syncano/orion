package serializers

import (
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database/fields"
)

type ChangeResponse struct {
	ID        int                    `json:"id"`
	Action    string                 `json:"status"`
	CreatedAt fields.Time            `json:"created_at"`
	Author    map[string]interface{} `json:"author"`
	Metadata  map[string]interface{} `json:"metadata"`
	Payload   map[string]interface{} `json:"payload"`
}

type ChangeSerializer struct{}

func (s ChangeSerializer) Response(i interface{}) interface{} {
	o := i.(*models.Change)

	return &ChangeResponse{
		ID:        o.ID,
		Action:    o.ActionString(),
		CreatedAt: fields.NewTime(&o.CreatedAt),
		Author:    o.Author,
		Metadata:  o.Metadata,
		Payload:   o.Payload,
	}
}
