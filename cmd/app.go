package cmd

import (
	_ "expvar" // Register expvar default http handler.
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis/v7"
	"github.com/urfave/cli/v2"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"

	"github.com/Syncano/orion/cmd/amqp"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/celery"
	"github.com/Syncano/orion/pkg/jobs"
	"github.com/Syncano/orion/pkg/log"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
	"github.com/Syncano/orion/pkg/version"
)

var (
	// App is the main structure of a cli application.
	App = cli.NewApp()

	dbOptions          = storage.DefaultDBOptions()
	dbInstancesOptions = storage.DefaultDBOptions()
	redisOptions       = redis.Options{}
	amqpChannel        *amqp.Channel

	jaegerExporter *jaeger.Exporter
)

func init() {
	App.Name = "Orion"
	App.Usage = "Application that enables running user provided unsecure code in a secure docker environment."
	App.Compiled = version.Buildtime
	App.Version = version.Current.String()
	App.Authors = []*cli.Author{
		{
			Name:  "Robert Kopaczewski",
			Email: "rk@23doors.com",
		},
	}
	App.Copyright = "(c) 2018 Syncano"
	App.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name: "debug", Usage: "enable debug mode",
			EnvVars: []string{"DEBUG"},
		},
		&cli.IntFlag{
			Name: "port", Aliases: []string{"p"}, Usage: "port for expvar server",
			EnvVars: []string{"METRIC_PORT"}, Value: 9080,
		},

		// Database options.
		&cli.StringFlag{
			Name: "db-name", Usage: "database name",
			EnvVars: []string{"DB_NAME"}, Value: "syncano", Destination: &dbOptions.Database,
		},
		&cli.StringFlag{
			Name: "db-user", Usage: "database user",
			EnvVars: []string{"DB_USER"}, Value: "syncano", Destination: &dbOptions.User,
		},
		&cli.StringFlag{
			Name: "db-pass", Usage: "database password",
			EnvVars: []string{"DB_PASS"}, Value: "syncano", Destination: &dbOptions.Password,
		},
		&cli.StringFlag{
			Name: "db-addr", Usage: "database address",
			EnvVars: []string{"DB_ADDR"}, Value: "postgresql:5432", Destination: &dbOptions.Addr,
		},

		// Database instances options.
		&cli.StringFlag{
			Name: "db-instances-name", Usage: "database name",
			EnvVars: []string{"DB_INSTANCES_NAME", "DB_NAME"}, Value: "syncano", Destination: &dbInstancesOptions.Database,
		},
		&cli.StringFlag{
			Name: "db-instances-user", Usage: "database user",
			EnvVars: []string{"DB_INSTANCES_USER", "DB_USER"}, Value: "syncano", Destination: &dbInstancesOptions.User,
		},
		&cli.StringFlag{
			Name: "db-instances-pass", Usage: "database password",
			EnvVars: []string{"DB_INSTANCES_PASS", "DB_PASS"}, Value: "syncano", Destination: &dbInstancesOptions.Password,
		},
		&cli.StringFlag{
			Name: "db-instances-host", Usage: "database address",
			EnvVars: []string{"DB_INSTANCES_ADDR", "DB_ADDR"}, Value: "postgresql:5432", Destination: &dbInstancesOptions.Addr,
		},

		// Tracing options.
		&cli.StringFlag{
			Name: "jaeger-collector-endpoint", Usage: "jaeger collector endpoint",
			EnvVars: []string{"JAEGER_COLLECTOR_ENDPOINT"}, Value: "http://jaeger:14268/api/traces",
		},
		&cli.Float64Flag{
			Name: "tracing-sampling", Usage: "tracing sampling probability value",
			EnvVars: []string{"TRACING_SAMPLING"}, Value: 0,
		},
		&cli.StringFlag{
			Name: "service-name", Aliases: []string{"n"}, Usage: "service name",
			EnvVars: []string{"SERVICE_NAME"}, Value: "orion",
		},

		// Redis options.
		&cli.StringFlag{
			Name: "redis-addr", Usage: "redis TCP address",
			EnvVars: []string{"REDIS_ADDR"}, Value: "redis:6379", Destination: &redisOptions.Addr,
		},

		// Broker options.
		&cli.StringFlag{
			Name: "broker-url", Usage: "amqp broker url",
			EnvVars: []string{"BROKER_URL"}, Value: "amqp://admin:mypass@rabbitmq//",
		},
	}
	App.Before = func(c *cli.Context) error {
		// Initialize random seed.
		rand.Seed(time.Now().UnixNano())

		numCPUs := runtime.NumCPU()
		runtime.GOMAXPROCS(numCPUs + 1) // numCPUs hot threads + one for async tasks.

		// Initialize logging.
		if err := sentry.Init(sentry.ClientOptions{}); err != nil {
			return err
		}

		if err := log.Init(c.Bool("debug"), sentry.CurrentHub().Client()); err != nil {
			return err
		}

		if c.Bool("debug") {
			settings.Common.Debug = true
		}

		// Serve expvar and checks.
		logger := log.Logger()
		logger.With(zap.Int("port", c.Int("port"))).Info("Serving http for expvar and checks")

		go func() {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", c.Int("port")), nil); err != nil && err != http.ErrServerClosed {
				logger.With(zap.Error(err)).Fatal("Serve error")
			}
		}()

		// Setup prometheus handler.
		exporter, err := prometheus.NewExporter(prometheus.Options{})
		if err != nil {
			logger.With(zap.Error(err)).Fatal("Prometheus exporter misconfiguration")
		}

		var views []*view.View
		views = append(views, ochttp.DefaultClientViews...)
		views = append(views, ochttp.DefaultServerViews...)
		views = append(views, ocgrpc.DefaultClientViews...)
		views = append(views, ocgrpc.DefaultServerViews...)

		if err := view.Register(views...); err != nil {
			logger.With(zap.Error(err)).Fatal("Opencensus views registration failed")
		}

		// Serve prometheus metrics.
		http.Handle("/metrics", exporter)

		// Initialize tracing.
		jaegerExporter, err = jaeger.NewExporter(jaeger.Options{
			CollectorEndpoint: c.String("jaeger-collector-endpoint"),
			Process: jaeger.Process{
				ServiceName: c.String("service-name"),
			},
			OnError: func(err error) {
				logger.With(zap.Error(err)).Warn("Jaeger tracing error")
			},
		})
		if err != nil {
			logger.With(zap.Error(err)).Fatal("Jaeger exporter misconfiguration")
		}

		trace.RegisterExporter(jaegerExporter)
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.ProbabilitySampler(c.Float64("tracing-sampling")),
		})

		// Initialize database client.
		storage.InitDB(dbOptions, dbInstancesOptions, c.Bool("debug"))

		// Validate default storage client.
		storage.Default()

		// Initialize redis client.
		storage.InitRedis(&redisOptions)
		cache.Init(storage.Redis())

		// Initialize AMQP client and celery wrapper.
		amqpChannel = new(amqp.Channel)
		if err := amqpChannel.Init(c.String("broker-url")); err != nil {
			return err
		}

		celery.Init(amqpChannel)

		return nil
	}
	App.After = func(c *cli.Context) error {
		// Redis teardown.
		storage.Redis().Close()

		// Database teardown.
		storage.DB().Close()

		// Sync loggers.
		log.Sync()

		// Shutdown AMQP client.
		amqpChannel.Shutdown()

		// Shutdown job system.
		jobs.Shutdown()

		// Close tracing reporter.
		jaegerExporter.Flush()

		// Flush remaining sentry events.
		sentry.Flush(5 * time.Second)

		return nil
	}
}
