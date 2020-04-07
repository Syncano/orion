package serializers

import (
	"net/http"

	"github.com/labstack/echo/v4"

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

type SocketTraceListResponse struct {
	ID         int                    `json:"id"`
	Status     string                 `json:"status"`
	ExecutedAt models.Time            `json:"executed_at"`
	Duration   *int                   `json:"duration"`
	Meta       map[string]interface{} `json:"meta"`
}

type SocketTraceListSerializer struct{}

func (s SocketTraceListSerializer) Response(i interface{}) interface{} {
	o := i.(*models.SocketTrace)

	var dur *int
	if o.Duration > 0 {
		dur = &o.Duration
	}

	return &SocketTraceListResponse{
		ID:         o.ID,
		Status:     o.Status,
		ExecutedAt: models.NewTime(&o.ExecutedAt),
		Duration:   dur,
		Meta:       o.Meta,
	}
}

type SocketTraceResponse struct {
	*SocketTraceListResponse

	Result map[string]interface{} `json:"result"`
	Args   map[string]interface{} `json:"args"`
}

type SocketTraceSerializer struct{}

func (s SocketTraceSerializer) Response(i interface{}) interface{} {
	o := i.(*models.SocketTrace)

	return &SocketTraceResponse{
		SocketTraceListResponse: SocketTraceListSerializer{}.Response(i).(*SocketTraceListResponse),
		Result:                  o.Result,
		Args:                    o.Args,
	}
}

type SocketTraceRenderResponse struct {
	ID         int                    `json:"id"`
	Status     string                 `json:"status"`
	ExecutedAt models.Time            `json:"executed_at"`
	Duration   *int                   `json:"duration"`
	Result     map[string]interface{} `json:"result"`
}

func (s SocketTraceSerializer) Render(c echo.Context, i interface{}) error {
	trace := i.(*models.SocketTrace)

	// Process raw response.
	if r, ok := trace.Result["response"]; ok {
		res := r.(map[string]interface{})

		if res["headers"] != nil {
			h := c.Response().Header()

			for k, v := range res["headers"].(map[string]string) {
				h.Set(k, v)
			}
		}

		return c.Blob(res["status"].(int), res["content_type"].(string), res["content"].([]byte))
	}

	var dur *int
	if trace.Duration > 0 {
		dur = &trace.Duration
	}

	return api.Render(c, statusToHTTPCode[trace.Status], &SocketTraceRenderResponse{
		ID:         trace.ID,
		Status:     trace.Status,
		ExecutedAt: models.NewTime(&trace.ExecutedAt),
		Duration:   dur,
		Result:     trace.Result,
	})
}
