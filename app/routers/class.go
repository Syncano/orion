package routers

import (
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/controllers"
)

// ClassRegister registers class routes.
func ClassRegister(r *echo.Group, m *middlewares) {
	g := r.Group("", m.Get()...)

	// List routes.
	g.GET("/", controllers.ClassList)
	g.POST("/", controllers.ClassCreate)

	// Detail routes.
	d := g.Group("/:class_name")
	d.GET("/", controllers.ClassRetrieve)
	d.PATCH("/", controllers.ClassUpdate)
	d.DELETE("/", controllers.ClassDelete)

	// Sub routes.
	sub := r.Group("/:class_name")
	DataObjectRegister(sub.Group("/objects"), m.Add(controllers.ClassContext))
}
