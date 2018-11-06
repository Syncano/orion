package routers

import (
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/controllers"
)

// InstanceRegister registers instance routes.
func InstanceRegister(r *echo.Group, m *middlewares) {
	g := r.Group("", m.Get()...)

	// List routes.
	g.GET("/", controllers.InstanceList)
	g.POST("/", controllers.InstanceCreate)

	// Detail routes.
	d := g.Group("/:instance_name")
	d.GET("/", controllers.InstanceRetrieve)
	d.PATCH("/", controllers.InstanceUpdate)
	d.DELETE("/", controllers.InstanceDelete)

	// Sub routes.
	sub := r.Group("/:instance_name")
	m = m.Add(controllers.InstanceContext, controllers.InstanceSubscriptionContext, controllers.BillingCheck).
		AddAuth(controllers.InstanceAuth)

	ClassRegister(sub.Group("/classes"), m)
	UserRegister(sub.Group("/users"), m)
	UserGroupRegister(sub.Group("/groups"), m)
}
