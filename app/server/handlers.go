package server

import (
	"fmt"
	"net/http"
	"time"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/lithammer/shortuuid"
	"go.uber.org/zap"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/pkg/log"
	"github.com/Syncano/orion/pkg/util"
)

const (
	ContextRequestID = "req_id"
)

func logg(c echo.Context, start time.Time, path string, l *zap.Logger, err error) {
	res := c.Response()
	code := res.Status
	end := time.Now()
	latency := end.Sub(start)
	dataLength := res.Size

	if dataLength < 0 {
		dataLength = 0
	}

	l = l.With(zap.Int("code", code),
		zap.Int64("len", dataLength),
		zap.String("ip", c.RealIP()),
		zap.Duration("took", latency),
	)
	msg := c.Request().Method + " " + path

	if err != nil {
		l = l.With(zap.Error(err))
	}

	if code > 499 && !util.IsCancellation(err) {
		l.Error(msg)
	} else {
		l.Info(msg)
	}
}

// Recovery recovers panics and logs them to sentry and zap.
func Recovery(logger *log.Logger) echo.MiddlewareFunc {
	l := logger.RawLogger()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			path := req.URL.Path

			defer func() {
				if rval := recover(); rval != nil {
					rvalStr := fmt.Sprint(rval)

					c.Error(api.NewGenericError(http.StatusInternalServerError, "Internal server error."))

					logg(c, start, path,
						l.With(zap.String("panic", rvalStr), zap.String("reqID", c.Get(ContextRequestID).(string))), nil)
				}
			}()

			return next(c)
		}
	}
}

// Logger logs requests using zap.
func Logger(logger *log.Logger) echo.MiddlewareFunc {
	l := logger.RawLogger()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			path := req.URL.Path

			err := next(c)

			if err != nil {
				c.Error(err)

				// Send to sentry on status >499.
				if c.Response().Status > 499 && !util.IsCancellation(err) {
					hub := sentryecho.GetHubFromContext(c)
					hub.CaptureException(err)
				}
			}

			logg(c, start, path,
				l, err)

			return nil
		}
	}
}

const RequestIDHeader = "X-Request-Id"

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = shortuuid.New()
			}

			c.Set(ContextRequestID, requestID)

			return next(c)
		}
	}
}
