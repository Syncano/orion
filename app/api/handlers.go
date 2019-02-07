package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis_rate"
	"github.com/labstack/echo"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/validators"
	"github.com/Syncano/orion/pkg/settings"
)

// Context keys.
const (
	ContextRealBodyKey  = "real_body"
	ContextRateLimitKey = "rate_limit"
)

var defaultErrors = map[int]string{
	http.StatusNotFound:              "Page not found",
	http.StatusMethodNotAllowed:      "Method not allowed",
	http.StatusRequestEntityTooLarge: "Request size limit exceeded",
}

// HTTPErrorHandler is a custom error handler for API.
func HTTPErrorHandler(err error, c echo.Context) {
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

// DisableBody ...
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

// RateLimit handles rate limit.
func RateLimit(limiter *redis_rate.Limiter, rateKey string, rateDur time.Duration, anonRateLimit bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authAdmin := c.Get(settings.ContextAdminKey)
			authAPIKey := c.Get(settings.ContextAPIKeyKey)
			if anonRateLimit && authAdmin == nil && authAPIKey == nil {
				// If we are to check against anon rates and admin is not logged in - check against anon rate limit.
				if _, delay, allowed := limiter.Allow(c.RealIP(), settings.API.AnonRateLimitS, time.Second); !allowed {
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

					if _, delay, allowed := limiter.Allow("a="+strconv.Itoa(admin.ID), settings.API.AdminRateLimitS, time.Second); !allowed {
						return rateLimitError(c, delay)
					}
				}
			} else {
				// If we are in instance scope - check against instance limits.
				var rate int64
				if rateKey == "" {
					rateKey = ContextRateLimitKey
				}
				if v := c.Get(rateKey); v != nil {
					rate = int64(v.(int))
				} else {
					rate = settings.API.InstanceRateLimitS
				}

				if _, delay, allowed := limiter.Allow("i="+strconv.Itoa(instance.(*models.Instance).ID), rate, rateDur); !allowed {
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
