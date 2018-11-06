package routers

import (
	"github.com/labstack/echo"
)

// Register registers all routes.
func Register(e *echo.Echo) {
	V3Register(e, e.Group("/v3"))
}
