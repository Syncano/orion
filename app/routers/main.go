package routers

import (
	"github.com/labstack/echo/v4"
)

// Register registers all routes.
func Register(e *echo.Echo) {
	V3Register(e, e.Group("/v3"))
}
