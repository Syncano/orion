package server

import (
	"context"
	"net"
	"net/http"
	"time"

	// using pprof when debug is true
	_ "net/http/pprof" // nolint:gosec

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/controllers"
	"github.com/Syncano/orion/app/routers"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/app/validators"
	"github.com/Syncano/pkg-go/v2/celery"
	"github.com/Syncano/pkg-go/v2/database"
	echo_middleware "github.com/Syncano/pkg-go/v2/echo_middleware"
	"github.com/Syncano/pkg-go/v2/log"
	"github.com/Syncano/pkg-go/v2/rediscache"
	"github.com/Syncano/pkg-go/v2/rediscli"
	"github.com/Syncano/pkg-go/v2/storage"
)

// Server defines a Web server wrapper.
type Server struct {
	srv   *http.Server
	ctr   *controllers.Controller
	log   *log.Logger
	debug bool
}

// NewServer initializes new Web server.
func NewServer(db *database.DB, fs *storage.Storage, redis *rediscli.Redis, rc *rediscache.Cache, cel *celery.Celery, logger *log.Logger, debug bool) (*Server, error) {
	ctr, err := controllers.New(db, fs, redis, rc, cel, logger)
	if err != nil {
		return nil, err
	}

	stdlog, _ := zap.NewStdLogAt(logger.Logger(), zap.WarnLevel)
	s := &Server{
		srv: &http.Server{
			ReadTimeout:  1 * time.Minute,
			WriteTimeout: 6 * time.Minute,
			IdleTimeout:  2 * time.Minute,
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
	// Top-down middlewares
	e.Use(
		echo_middleware.RequestID(),
		middleware.CORSWithConfig(middleware.CORSConfig{MaxAge: 86400}),
		echo_middleware.OpenCensus(),
		sentryecho.New(sentryecho.Options{
			Repanic: true,
		}),
		echo_middleware.Logger(s.log),
		echo_middleware.Recovery(s.log),
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
