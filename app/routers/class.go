package routers

import (
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/controllers"
)

// ClassRegister registers class routes.
func ClassRegister(ctr *controllers.Controller, r *echo.Group, m *middlewares) {
	g := r.Group("", m.Get(ctr)...)

	// List routes.
	g.GET("/", ctr.ClassList)
	g.POST("/", ctr.ClassCreate)

	// Detail routes.
	d := g.Group("/:class_name")
	d.GET("/", ctr.ClassRetrieve)
	d.PATCH("/", ctr.ClassUpdate)
	d.DELETE("/", ctr.ClassDelete)

	// Sub routes.
	sub := r.Group("/:class_name")
	DataObjectRegister(ctr, sub.Group("/objects"), m.Add(ctr.ClassContext))
}
