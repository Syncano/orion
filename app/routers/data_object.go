package routers

import (
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/controllers"
)

// DataObjectRegister registers data object routes.
func DataObjectRegister(r *echo.Group, m *middlewares) {
	g := r.Group("", m.Get()...)

	// List routes.
	g.GET("/", controllers.DataObjectList)
	g.POST("/", controllers.DataObjectCreate)

	// Detail routes.
	d := g.Group("/:object_id")
	d.GET("/", controllers.DataObjectRetrieve)
	d.PATCH("/", controllers.DataObjectUpdate)
	d.DELETE("/", controllers.DataObjectDelete)
}
