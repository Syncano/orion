package routers

import (
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/controllers"
)

// InstanceRegister registers instance routes.
func InstanceRegister(ctr *controllers.Controller, r *echo.Group, m *middlewares) {
	g := r.Group("", m.Get(ctr)...)

	// List routes.
	g.GET("/", ctr.InstanceList)
	g.POST("/", ctr.InstanceCreate)

	// Detail routes.
	d := g.Group("/:instance_name")
	d.GET("/", ctr.InstanceRetrieve)
	d.PATCH("/", ctr.InstanceUpdate)
	d.DELETE("/", ctr.InstanceDelete)

	// Sub routes.
	sub := r.Group("/:instance_name")
	m = m.Add(ctr.InstanceContext, ctr.InstanceSubscriptionContext, ctr.BillingCheck).
		AddAuth(ctr.InstanceAuth)

	ClassRegister(ctr, sub.Group("/classes"), m)
	UserRegister(ctr, sub.Group("/users"), m)
	UserGroupRegister(ctr, sub.Group("/groups"), m)
	SocketEndpointRegister(ctr, sub.Group("/endpoints/sockets"), m)
}
