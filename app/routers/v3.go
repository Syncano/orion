package routers

import (
	"github.com/go-redis/redis_rate"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/bytes"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/controllers"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
)

// V3Register registers v3 routes.
func V3Register(e *echo.Echo, g *echo.Group) {
	InstanceRegister(g.Group("/instances"), standardMiddlewares(e))
	g.POST("/cache_invalidate/", controllers.CacheInvalidate)
}

type middlewares struct {
	e         *echo.Echo
	limiter   *redis_rate.Limiter
	chain     []echo.MiddlewareFunc
	authChain []echo.MiddlewareFunc

	Auth         bool
	AuthUser     bool
	RequireAuth  bool
	RequireAdmin bool
	RequireUser  bool

	InstanceRateLimit *settings.RateData
	AdminRateLimit    *settings.RateData
	AnonRateLimit     *settings.RateData

	DisableBody bool
	SizeLimit   int64
}

func (m *middlewares) Get() []echo.MiddlewareFunc {
	f := m.chain

	if m.DisableBody {
		f = append(f, api.DisableBody)
	}

	if m.SizeLimit > 0 {
		f = append(f, middleware.BodyLimit(bytes.Format(m.SizeLimit)))
	}

	f = append(f, api.MethodOverride)

	if m.Auth {
		f = append(f, controllers.Auth)

		if m.AuthUser {
			f = append(f, controllers.AuthUser)
		}

		if m.RequireAuth {
			if m.RequireAdmin {
				f = append(f, controllers.RequireAdmin)
			} else {
				f = append(f, controllers.RequireAPIKeyOrAdmin)
			}

			if m.RequireUser {
				f = append(f, controllers.RequireUser)
			}

			f = append(f, m.authChain...)
		}
	}

	f = append(f, api.RateLimit(m.limiter, m.InstanceRateLimit, m.AdminRateLimit, m.AnonRateLimit))

	return f
}

func (m *middlewares) Add(f ...echo.MiddlewareFunc) *middlewares {
	m = m.Copy()
	m.chain = append(m.chain, f...)

	return m
}

func (m *middlewares) AddAuth(f ...echo.MiddlewareFunc) *middlewares {
	m = m.Copy()
	m.authChain = append(m.authChain, f...)

	return m
}

func (m *middlewares) Copy() *middlewares {
	cp := &middlewares{}
	*cp = *m
	cp.chain = append([]echo.MiddlewareFunc(nil), cp.chain...)
	cp.authChain = append([]echo.MiddlewareFunc(nil), cp.authChain...)

	return cp
}

func standardMiddlewares(e *echo.Echo) *middlewares {
	return &middlewares{
		e:       e,
		limiter: redis_rate.NewLimiter(storage.Redis()),

		Auth:         true,
		AuthUser:     true,
		RequireAuth:  true,
		RequireAdmin: true,

		InstanceRateLimit: settings.API.InstanceRateLimit,
		AdminRateLimit:    settings.API.AdminRateLimit,
		AnonRateLimit:     settings.API.AnonRateLimit,

		DisableBody: false,
		SizeLimit:   settings.API.MaxPayloadSize,
	}
}
