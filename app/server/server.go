package server

import (
	"context"
	"net"
	"net/http"
	"time"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/controllers"
	"github.com/Syncano/orion/app/routers"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/app/validators"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/celery"
	"github.com/Syncano/orion/pkg/log"
	"github.com/Syncano/orion/pkg/storage"
)

// Server defines a Web server wrapper.
type Server struct {
	srv   *http.Server
	ctr   *controllers.Controller
	log   *log.Logger
	debug bool
}

// NewServer initializes new Web server.
func NewServer(db *storage.Database, fs *storage.Storage, redis *storage.Redis, c *cache.Cache, cel *celery.Celery, logger *log.Logger, debug bool) (*Server, error) {
	ctr, err := controllers.New(db, fs, redis, c, cel, logger)
	if err != nil {
		return nil, err
	}

	stdlog, _ := zap.NewStdLogAt(logger.Logger(), zap.WarnLevel)
	s := &Server{
		srv: &http.Server{
			ReadTimeout:  6 * time.Minute,
			WriteTimeout: 6 * time.Minute,
			ErrorLog:     stdlog,
		},
		debug: debug,
		ctr:   ctr,
		log:   logger,
	}
	s.srv.Handler = s.setupRouter()

	return s, nil
}

func (s *Server) setupRouter() *echo.Echo {
	e := echo.New()
	// Bottom up middlewares
	e.Use(
		Recovery(s.log),
		Logger(s.log),
		sentryecho.New(sentryecho.Options{
			Repanic: true,
		}),
		middleware.CORSWithConfig(middleware.CORSConfig{MaxAge: 86400}),
		OpenCensus(),
	)

	// Register profiling if debug is on.
	// go tool pprof http://.../debug/pprof/profile
	// go tool pprof http://.../debug/pprof/heap
	if s.debug {
		e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))
	}

	// If MediaPrefix is set to local files - serve them.
	if settings.API.StorageURL[0] == '/' {
		e.Static(settings.API.StorageURL[:len(settings.API.StorageURL)-1], "media")
	}

	e.HTTPErrorHandler = api.HTTPErrorHandler
	e.Binder = &api.Binder{}
	e.Validator = validators.NewValidator()

	routers.Register(s.ctr, e)

	return e
}

// Serve handles requests on incoming connections.
func (s *Server) Serve(lis net.Listener) error {
	return s.srv.Serve(lis)
}

// Shutdown stops gracefully server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
