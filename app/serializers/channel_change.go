package serializers

import (
	"github.com/Syncano/orion/app/models"
)

// ChangeResponse ...
type ChangeResponse struct {
	ID        int                    `json:"id"`
	Action    string                 `json:"status"`
	CreatedAt *models.Time           `json:"created_at"`
	Author    map[string]interface{} `json:"author"`
	Metadata  map[string]interface{} `json:"metadata"`
	Payload   map[string]interface{} `json:"payload"`
}

// ChangeSerializer ...
type ChangeSerializer struct{}

// Response ...
func (s ChangeSerializer) Response(i interface{}) interface{} {
	o := i.(*models.Change)

	var t *models.Time
	if !o.CreatedAt.IsZero() {
		nt := models.NewTime(&o.CreatedAt)
		t = &nt
	}

	return &ChangeResponse{
		ID:        o.ID,
		Action:    o.ActionString(),
		CreatedAt: t,
		Author:    o.Author,
		Metadata:  o.Metadata,
		Payload:   o.Payload,
	}
}
