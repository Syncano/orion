package routers

import (
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/controllers"
)

// DataObjectRegister registers data object routes.
func DataObjectRegister(ctr *controllers.Controller, r *echo.Group, m *middlewares) {
	g := r.Group("", m.Get(ctr)...)

	// List routes.
	g.GET("/", ctr.DataObjectList)
	g.POST("/", ctr.DataObjectCreate)

	// Detail routes.
	d := g.Group("/:object_id")
	d.GET("/", ctr.DataObjectRetrieve)
	d.PATCH("/", ctr.DataObjectUpdate)
	d.DELETE("/", ctr.DataObjectDelete)
}
