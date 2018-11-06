package server

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/pkg/log"
)

const (
	traceContextLines = 3
	traceSkipFrames   = 2
)

func logg(c echo.Context, start time.Time, path string, l *zap.Logger) {
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
	if code > 499 {
		l.Error(msg)
	} else {
		l.Info(msg)
	}
}

// Recovery recovers panics and logs them to sentry and zap.
func Recovery() echo.MiddlewareFunc {
	logger := log.RawLogger()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			path := req.URL.Path

			defer func() {
				if rval := recover(); rval != nil {
					flags := map[string]string{
						"endpoint": req.RequestURI,
					}

					rvalStr := fmt.Sprint(rval)
					raven.CaptureMessage(rvalStr, flags, raven.NewException(errors.New(rvalStr),
						raven.NewStacktrace(traceSkipFrames, traceContextLines, nil)),
						raven.NewHttp(req))
					c.Error(api.NewGenericError(http.StatusInternalServerError, "Internal server error."))
					logg(c, start, path, logger.With(zap.String("panic", rvalStr)))
				}
			}()

			return next(c)
		}
	}
}

// Logger logs requests using zap.
func Logger() echo.MiddlewareFunc {
	logger := log.RawLogger()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			path := req.URL.Path
			l := logger

			if err := next(c); err != nil {
				c.Error(err)

				// Send to sentry on status >499.
				if c.Response().Status > 499 {
					flags := map[string]string{
						"endpoint": req.RequestURI,
					}

					raven.CaptureMessage(err.Error(), flags, raven.NewException(err,
						raven.NewStacktrace(traceSkipFrames, traceContextLines, nil)),
						raven.NewHttp(req))
				}
				l = logger.With(zap.Error(err))
			}
			logg(c, start, path, l)
			return nil
		}
	}
}
