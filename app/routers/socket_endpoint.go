package routers

import (
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/controllers"
	"github.com/Syncano/orion/pkg/settings"
)

// SocketEndpointRegister registers socket endpoint routes.
func SocketEndpointRegister(r *echo.Group, m *middlewares) {
	m = m.Copy()
	m.RequireAuth = false
	m.AnonRateLimit = &settings.RateData{Limit: -1}
	m.SizeLimit = settings.Socket.MaxPayloadSize
	g := r.Group("", m.Get()...)

	// List all socket endpoints.
	g.GET("/", controllers.SocketEndpointList)

	// List socket endpoints by socket name.
	d := g.Group("/:socket_name")
	d.GET("/", controllers.SocketEndpointList)

	// Socket endpoint routes.
	d = d.Group("/*")
	d.Any("/", controllers.SocketEndpointMap)
}
