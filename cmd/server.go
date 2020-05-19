package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Syncano/orion/app/server"
	"github.com/Syncano/orion/pkg/celery"
	"github.com/Syncano/orion/pkg/version"
)

var serverCmd = &cli.Command{
	Name:        "server",
	Usage:       "Starts orion server.",
	Description: `Orion server provides v3+ API for Syncano platform.`,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name: "port", Usage: "port for web server",
			EnvVars: []string{"PORT"}, Value: 8000,
		},
	},
	Action: func(c *cli.Context) error {
		logg := logger.Logger()

		logg.With(
			zap.String("version", App.Version),
			zap.String("gitsha", version.GitSHA),
			zap.Time("buildtime", App.Compiled),
		).Info("Server starting")

		// Create new http server.
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int("port")))
		if err != nil {
			return err
		}
		srv, err := server.NewServer(db,
			fs,
			rredis,
			cach,
			celery.New(amqpChannel),
			logger,
			c.Bool("debug"))
		if err != nil {
			return err
		}
		go func() {
			if err := srv.Serve(lis); err != nil && err != http.ErrServerClosed {
				logg.With(zap.Error(err)).Fatal("Serve error")
			}
		}()
		logg.With(zap.Int("port", c.Int("port"))).Info("Serving web")

		// Setup health check.
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		// Handle SIGINT and SIGTERM.
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch

		// Graceful shutdown.
		logg.Info("Shutting down")
		srv.Shutdown(context.Background()) // nolint: errcheck
		return nil
	},
}

func init() {
	App.Commands = append(App.Commands, serverCmd)
}
