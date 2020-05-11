package routers

import (
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/controllers"
)

// Register registers all routes.
func Register(ctr *controllers.Controller, e *echo.Echo) {
	V3Register(ctr, e, e.Group("/v3"))
}
