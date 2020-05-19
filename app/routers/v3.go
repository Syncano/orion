package routers

import (
	"github.com/go-redis/redis/v7"
	"github.com/go-redis/redis_rate/v7"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/bytes"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/controllers"
	"github.com/Syncano/orion/app/settings"
)

// V3Register registers v3 routes.
func V3Register(ctr *controllers.Controller, e *echo.Echo, g *echo.Group) {
	InstanceRegister(ctr, g.Group("/instances"), standardMiddlewares(e, ctr.Redis().Client()))
	g.POST("/cache_invalidate/", ctr.CacheInvalidate)
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

func (m *middlewares) Get(ctr *controllers.Controller) []echo.MiddlewareFunc {
	f := m.chain

	if m.DisableBody {
		f = append(f, api.DisableBody)
	}

	if m.SizeLimit > 0 {
		f = append(f, middleware.BodyLimit(bytes.Format(m.SizeLimit)))
	}

	f = append(f, api.MethodOverride)

	if m.Auth {
		f = append(f, ctr.Auth)

		if m.AuthUser {
			f = append(f, ctr.AuthUser)
		}

		if m.RequireAuth {
			if m.RequireAdmin {
				f = append(f, ctr.RequireAdmin)
			} else {
				f = append(f, ctr.RequireAPIKeyOrAdmin)
			}

			if m.RequireUser {
				f = append(f, ctr.RequireUser)
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

func standardMiddlewares(e *echo.Echo, r *redis.Client) *middlewares {
	return &middlewares{
		e:       e,
		limiter: redis_rate.NewLimiter(r),

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
