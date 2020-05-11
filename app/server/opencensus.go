package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opencensus.io/plugin/ochttp"
)

// OpenCensusConfig defines the config for OpenCensus middleware.
type OpenCensusConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper
}

// DefaultOpenCensusConfig is the default OpenCensus middleware config.
var DefaultOpenCensusConfig = OpenCensusConfig{
	Skipper: middleware.DefaultSkipper,
}

// OpenCensus returns a middleware that collect HTTP requests and response
// metrics.
func OpenCensus() echo.MiddlewareFunc {
	return OpenCensusWithConfig(DefaultOpenCensusConfig)
}

// OpenCensusWithConfig returns a OpenCensus middleware with config.
// See: `OpenCensus()`.
func OpenCensusWithConfig(cfg OpenCensusConfig) echo.MiddlewareFunc {
	// Defaults
	if cfg.Skipper == nil {
		cfg.Skipper = DefaultOpenCensusConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if cfg.Skipper(c) {
				return next(c)
			}

			handler := &ochttp.Handler{
				FormatSpanName: func(r *http.Request) string {
					return c.Path()
				},
				Handler: http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						c.SetRequest(r)
						c.SetResponse(echo.NewResponse(w, c.Echo()))
						err = next(c)
					},
				),
			}

			handler.ServeHTTP(c.Response(), c.Request())

			return
		}
	}
}
