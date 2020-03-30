package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/Syncano/orion/app/proto/codebox"
	"github.com/Syncano/orion/app/proto/codebox/broker"
	"github.com/Syncano/orion/pkg/log"
	"github.com/Syncano/orion/pkg/server"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/version"
)

var serverCmd = cli.Command{
	Name:  "server",
	Usage: "Starts server to serve as a front for load balancers.",
	Description: `Servers pass workload in correct way to available load balancers.
As there is no authentication, always run it in a private network.`,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name: "port", Usage: "port for web server",
			EnvVar: "PORT", Value: 80,
		},

		cli.StringFlag{
			Name: "codebox-addr", Usage: "addr for codebox broker server",
			EnvVar: "CODEBOX_ADDR", Value: "codebox-broker:80",
		},
	},
	Action: func(c *cli.Context) error {
		logger := log.Logger()

		logger.With(
			zap.String("version", App.Version),
			zap.String("gitsha", version.GitSHA),
			zap.Time("buildtime", App.Compiled),
		).Info("Server starting")

		conn, err := grpc.Dial(c.String("codebox-addr"), settings.DefaultGRPCDialOptions...)
		if err != nil {
			return err
		}
		codebox.Runner = broker.NewScriptRunnerClient(conn)

		// Create new http server.
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int("port")))
		if err != nil {
			return err
		}
		srv, err := server.NewServer(c.GlobalBool("debug"))
		if err != nil {
			return err
		}
		go func() {
			if err := srv.Serve(lis); err != nil && err != http.ErrServerClosed {
				logger.With(zap.Error(err)).Fatal("Serve error")
			}
		}()
		logger.With(zap.Int("port", c.Int("port"))).Info("Serving web")

		// Setup health check.
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		// Handle SIGINT and SIGTERM.
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch

		// Graceful shutdown.
		logger.Info("Shutting down")
		srv.Shutdown(context.Background()) // nolint: errcheck
		return nil
	},
}

func init() {
	App.Commands = append(App.Commands, serverCmd)
}
