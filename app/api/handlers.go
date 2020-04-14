package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis_rate/v7"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/validators"
	"github.com/Syncano/orion/pkg/settings"
)

// Context keys.
const (
	ContextRealBodyKey          = "real_body"
	ContextInstanceRateLimitKey = "instance_rate_limit"
	ContextAdminRateLimitKey    = "admin_rate_limit"
	ContextAnonRateLimitKey     = "anon_rate_limit"
)

var defaultErrors = map[int]string{
	http.StatusNotFound:              "Page not found",
	http.StatusMethodNotAllowed:      "Method not allowed",
	http.StatusRequestEntityTooLarge: "Request size limit exceeded",
}

// HTTPErrorHandler is a custom error handler for API.
func HTTPErrorHandler(err error, c echo.Context) {
	// context.Canceled on request means that http/2.0 client disconnected.
	if c.Request().Context().Err() == context.Canceled {
		_ = c.NoContent(http.StatusNoContent)
		return
	}

	// Process API errors.
	if e, ok := err.(*Error); ok {
		Render(c, e.Code, e.Data) // nolint: errcheck
		return
	}

	// Process echo errors.
	if e, ok := err.(*echo.HTTPError); ok {
		message := e.Message
		if m, ok := defaultErrors[e.Code]; ok {
			message = m
		}

		Render(c, e.Code, map[string]interface{}{"detail": fmt.Sprintf("%s.", message)}) // nolint: errcheck

		return
	}

	// Process validation errors.
	if e, ok := err.(validator.ValidationErrors); ok {
		Render(c, http.StatusBadRequest, validators.TranslateErrors(e)) // nolint: errcheck
		return
	}

	Render(c, http.StatusInternalServerError, map[string]string{"detail": "Internal server error."}) // nolint: errcheck
}

type empty struct{}

type fakeBody struct{}

func (f fakeBody) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (f fakeBody) Close() error {
	return nil
}

func DisableBody(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		c.Set(ContextRealBodyKey, req.Body)
		req.Body = fakeBody{}

		return next(c)
	}
}

func rateLimitError(c echo.Context, delay time.Duration) error {
	c.Response().Header().Set("Retry-After", strconv.Itoa(int(delay.Seconds())))
	return NewGenericError(http.StatusTooManyRequests, "Too many requests.")
}

func checkLimit(limiter *redis_rate.Limiter, name string, rate *settings.RateData) (delay time.Duration, allowed bool) {
	if rate.Limit < 0 {
		allowed = true
		return
	}

	_, delay, allowed = limiter.Allow(name, rate.Limit, rate.Duration)

	return
}

// RateLimit handles rate limit.
func RateLimit(limiter *redis_rate.Limiter, instanceRateLimit, adminRateLimit, anonRateLimit *settings.RateData) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authAdmin := c.Get(settings.ContextAdminKey)
			authAPIKey := c.Get(settings.ContextAPIKeyKey)

			if authAdmin == nil && authAPIKey == nil {
				// If admin is not logged in - check against anon rps limit.
				var rate *settings.RateData

				switch v := c.Get(ContextAnonRateLimitKey); {
				case v != nil:
					rate = v.(*settings.RateData)
				case anonRateLimit != nil:
					rate = anonRateLimit
				default:
					rate = settings.API.AnonRateLimit
				}

				if delay, allowed := checkLimit(limiter, c.RealIP(), rate); !allowed {
					return rateLimitError(c, delay)
				}
				// No reason to check other cases.
				// Anon rate limit should be explicitly disabled on instance endpoints with no auth.
				return next(c)
			}

			instance := c.Get(settings.ContextInstanceKey)
			if instance == nil {
				// If admin is logged in but outside of instance scope - check against admin rate limit.
				if authAdmin != nil {
					admin := authAdmin.(*models.Admin)
					// Allow staff users to not be limited.
					if admin.IsStaff {
						return next(c)
					}

					var rate *settings.RateData

					switch v := c.Get(ContextAdminRateLimitKey); {
					case v != nil:
						rate = v.(*settings.RateData)
					case adminRateLimit != nil:
						rate = adminRateLimit
					default:
						rate = settings.API.AdminRateLimit
					}

					if delay, allowed := checkLimit(limiter, "a="+strconv.Itoa(admin.ID), rate); !allowed {
						return rateLimitError(c, delay)
					}
				}
			} else {
				// If we are in instance scope - check against instance limits.
				var rate *settings.RateData

				switch v := c.Get(ContextInstanceRateLimitKey); {
				case v != nil:
					rate = v.(*settings.RateData)
				case instanceRateLimit != nil:
					rate = instanceRateLimit
				default:
					rate = settings.API.InstanceRateLimit
				}

				if delay, allowed := checkLimit(limiter, "i="+strconv.Itoa(instance.(*models.Instance).ID), rate); !allowed {
					return rateLimitError(c, delay)
				}
			}

			return next(c)
		}
	}
}

var validMethods = map[string]empty{
	http.MethodGet:    {},
	http.MethodPost:   {},
	http.MethodPut:    {},
	http.MethodPatch:  {},
	http.MethodDelete: {},
}

// MethodOverride is a middleware that based on POST _method changes request's method.
func MethodOverride(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		if req.Method == http.MethodPost {
			m := c.FormValue("_method")
			if _, ok := validMethods[m]; ok {
				delete(req.Form, "_method")

				if m != req.Method {
					req.Method = m
					c.Echo().Router().Find(req.Method, req.URL.Path, c)

					return c.Handler()(c)
				}
			}
		}

		return next(c)
	}
}
