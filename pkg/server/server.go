package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/routers"
	"github.com/Syncano/orion/app/validators"
	"github.com/Syncano/orion/pkg/log"
	"github.com/Syncano/orion/pkg/settings"
)

// Server defines a Web server wrapper.
type Server struct {
	srv   *http.Server
	debug bool
}

// NewServer initializes new Web server.
func NewServer(debug bool) (*Server, error) {
	s := &Server{
		srv: &http.Server{
			ReadTimeout:  6 * time.Minute,
			WriteTimeout: 6 * time.Minute,
			ErrorLog:     zap.NewStdLog(log.Logger()),
		},
		debug: debug,
	}
	s.srv.Handler = s.setupRouter()

	return s, nil
}

func (s *Server) setupRouter() *echo.Echo {
	e := echo.New()
	e.Use(
		Recovery(),
		Logger(),
		middleware.CORSWithConfig(middleware.CORSConfig{MaxAge: 86400}),
		OpenTracing(),
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

	routers.Register(e)

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
