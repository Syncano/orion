package routers

import (
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/controllers"
	"github.com/Syncano/orion/app/settings"
)

// SocketEndpointRegister registers socket endpoint routes.
func SocketEndpointRegister(ctr *controllers.Controller, r *echo.Group, m *middlewares) {
	m = m.Copy()
	m.RequireAuth = false
	m.AnonRateLimit = &settings.RateData{Limit: -1}
	m.SizeLimit = settings.Socket.MaxPayloadSize
	g := r.Group("", m.Get(ctr)...)

	// List all socket endpoints.
	g.GET("/", ctr.SocketEndpointList)

	// List socket endpoints by socket name.
	d := g.Group("/:socket_name")
	d.GET("/", ctr.SocketEndpointList)

	// Socket endpoint routes.
	d = d.Group("/*")
	d.Any("/", ctr.SocketEndpointMap)
}
