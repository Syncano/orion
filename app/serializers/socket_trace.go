package serializers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
)

var (
	statusToHTTPCode = map[string]int{
		models.TraceStatusSuccess: http.StatusOK,
		models.TraceStatusFailure: http.StatusInternalServerError,
		models.TraceStatusBlocked: http.StatusTooManyRequests,
		models.TraceStatusTimeout: http.StatusRequestTimeout,
	}
)

// SocketTraceListResponse ...
type SocketTraceListResponse struct {
	ID         int                    `json:"id"`
	Status     string                 `json:"status"`
	ExecutedAt *models.Time           `json:"executed_at"`
	Duration   *int                   `json:"duration"`
	Meta       map[string]interface{} `json:"meta"`
}

// SocketTraceListSerializer ...
type SocketTraceListSerializer struct{}

// Response ...
func (s SocketTraceListSerializer) Response(i interface{}) interface{} {
	o := i.(*models.SocketTrace)
	var t *models.Time
	if !o.ExecutedAt.IsZero() {
		t = models.NewTime(&o.ExecutedAt)
	}

	var dur *int
	if o.Duration > 0 {
		dur = &o.Duration
	}
	return &SocketTraceListResponse{
		ID:         o.ID,
		Status:     o.Status,
		ExecutedAt: t,
		Duration:   dur,
		Meta:       o.Meta,
	}
}

// SocketTraceResponse ...
type SocketTraceResponse struct {
	*SocketTraceListResponse

	Result map[string]interface{} `json:"result"`
	Args   map[string]interface{} `json:"args"`
}

// SocketTraceSerializer ...
type SocketTraceSerializer struct{}

// Response ...
func (s SocketTraceSerializer) Response(i interface{}) interface{} {
	o := i.(*models.SocketTrace)
	return &SocketTraceResponse{
		SocketTraceListResponse: SocketTraceListSerializer{}.Response(i).(*SocketTraceListResponse),
		Result:                  o.Result,
		Args:                    o.Args,
	}
}

// SocketTraceRenderResponse ...
type SocketTraceRenderResponse struct {
	ID         int                    `json:"id"`
	Status     string                 `json:"status"`
	ExecutedAt *models.Time           `json:"executed_at"`
	Duration   *int                   `json:"duration"`
	Result     map[string]interface{} `json:"result"`
}

// Render ...
func (s SocketTraceSerializer) Render(c echo.Context, i interface{}) error {
	trace := i.(*models.SocketTrace)

	// Process raw response.
	if r, ok := trace.Result["response"]; ok {
		res := r.(map[string]interface{})

		if res["header"] != nil {
			h := c.Response().Header()
			for k, v := range res["header"].(map[string]string) {
				h.Set(k, v)
			}
		}
		return c.Blob(res["status"].(int), res["content_type"].(string), res["content"].([]byte))
	}

	// Process Trace Response.
	var t *models.Time
	if !trace.ExecutedAt.IsZero() {
		t = models.NewTime(&trace.ExecutedAt)
	}

	var dur *int
	if trace.Duration > 0 {
		dur = &trace.Duration
	}

	return api.Render(c, statusToHTTPCode[trace.Status], &SocketTraceRenderResponse{
		ID:         trace.ID,
		Status:     trace.Status,
		ExecutedAt: t,
		Duration:   dur,
		Result:     trace.Result,
	})
}
