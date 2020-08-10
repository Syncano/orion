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
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
	"google.golang.org/grpc/grpclog"

	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/app/version"
	"github.com/Syncano/orion/cmd/amqp"
	"github.com/Syncano/orion/pkg/jobs"
	"github.com/Syncano/pkg-go/v2/database"
	"github.com/Syncano/pkg-go/v2/log"
	"github.com/Syncano/pkg-go/v2/rediscache"
	"github.com/Syncano/pkg-go/v2/rediscli"
	"github.com/Syncano/pkg-go/v2/storage"
)

var (
	// App is the main structure of a cli application.
	App = cli.NewApp()

	dbOptions          = database.DefaultDBOptions()
	dbInstancesOptions = database.DefaultDBOptions()
	redisOptions       = redis.Options{}
	amqpChannel        *amqp.Channel

	jaegerExporter *jaeger.Exporter
	db             *database.DB
	fs             *storage.Storage
	storRedis      *rediscli.Redis
	cache          *rediscache.Cache
	logger         *log.Logger
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
	App.Copyright = "Syncano"
	App.Flags = []cli.Flag{
		&cli.StringFlag{
			Name: "log-level", Usage: "logging level",
			EnvVars: []string{"LOGLEVEL"}, Value: zapcore.InfoLevel.String(),
		},
		&cli.BoolFlag{
			Name: "debug", Usage: "enable debug mode",
			EnvVars: []string{"DEBUG"},
		},
		&cli.IntFlag{
			Name: "metrics-port", Aliases: []string{"mp"}, Usage: "port for metrics",
			EnvVars: []string{"METRICS_PORT"}, Value: 9080,
		},

		// Database options.
		&cli.StringFlag{
			Name: "db-name", Usage: "database name",
			EnvVars: []string{"DB_NAME", "PGDATABASE"}, Value: "syncano", Destination: &dbOptions.Database,
		},
		&cli.StringFlag{
			Name: "db-user", Usage: "database user",
			EnvVars: []string{"DB_USER", "PGUSER"}, Value: "syncano", Destination: &dbOptions.User,
		},
		&cli.StringFlag{
			Name: "db-pass", Usage: "database password",
			EnvVars: []string{"DB_PASS", "PGPASSWORD"}, Value: "syncano", Destination: &dbOptions.Password,
		},
		&cli.StringFlag{
			Name: "db-host", Usage: "database host",
			EnvVars: []string{"DB_HOST", "PGHOST"}, Value: "postgresql",
		},
		&cli.StringFlag{
			Name: "db-port", Usage: "database port",
			EnvVars: []string{"DB_PORT", "PGPORT"}, Value: "5432",
		},

		// Database instances options.
		&cli.StringFlag{
			Name: "db-instances-name", Usage: "instances database name",
			EnvVars: []string{"DB_INSTANCES_NAME", "DB_NAME", "PGDATABASE"}, Value: "syncano", Destination: &dbInstancesOptions.Database,
		},
		&cli.StringFlag{
			Name: "db-instances-user", Usage: "instances database user",
			EnvVars: []string{"DB_INSTANCES_USER", "DB_USER", "PGUSER"}, Value: "syncano", Destination: &dbInstancesOptions.User,
		},
		&cli.StringFlag{
			Name: "db-instances-pass", Usage: "instances database password",
			EnvVars: []string{"DB_INSTANCES_PASS", "DB_PASS", "PGPASSWORD"}, Value: "syncano", Destination: &dbInstancesOptions.Password,
		},
		&cli.StringFlag{
			Name: "db-instances-host", Usage: "instances database host",
			EnvVars: []string{"DB_INSTANCES_HOST", "DB_HOST", "PGHOST"}, Value: "postgresql",
		},
		&cli.StringFlag{
			Name: "db-instances-port", Usage: "instances database port",
			EnvVars: []string{"DB_INSTANCES_PORT", "DB_PORT", "PGPORT"}, Value: "5432",
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

		var err error
		if logger, err = log.New(sentry.CurrentHub().Client(),
			log.WithLogLevel(c.String("log_level")),
			log.WithDebug(c.Bool("debug")),
		); err != nil {
			return err
		}

		logg := logger.Logger()

		// Set grpc logger.
		var zapgrpcOpts []zapgrpc.Option
		if c.Bool("debug") {
			zapgrpcOpts = append(zapgrpcOpts, zapgrpc.WithDebug())
		}

		grpclog.SetLogger(zapgrpc.NewLogger(logg, zapgrpcOpts...)) // nolint: staticcheck

		if c.Bool("debug") {
			settings.Common.Debug = true
		}

		// Setup prometheus handler.
		exporter, err := prometheus.NewExporter(prometheus.Options{})
		if err != nil {
			logg.With(zap.Error(err)).Fatal("Prometheus exporter misconfiguration")
		}

		var views []*view.View
		views = append(views, ochttp.DefaultClientViews...)
		views = append(views, ochttp.DefaultServerViews...)
		views = append(views, ocgrpc.DefaultClientViews...)
		views = append(views, ocgrpc.DefaultServerViews...)

		if err := view.Register(views...); err != nil {
			logg.With(zap.Error(err)).Fatal("Opencensus views registration failed")
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
				logg.With(zap.Error(err)).Warn("Jaeger tracing error")
			},
		})
		if err != nil {
			logg.With(zap.Error(err)).Fatal("Jaeger exporter misconfiguration")
		}

		trace.RegisterExporter(jaegerExporter)
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.ProbabilitySampler(c.Float64("tracing-sampling")),
		})

		// Initialize database client.
		dbOptions.Addr = fmt.Sprintf("%s:%s", c.String("db-host"), c.String("db-port"))
		dbInstancesOptions.Addr = fmt.Sprintf("%s:%s", c.String("db-instances-host"), c.String("db-instances-port"))
		db = database.NewDB(dbOptions, dbInstancesOptions, logger, c.Bool("debug"))

		fs = storage.NewStorage(settings.Common.Location, settings.Buckets, settings.API.Host, settings.API.StorageURL)

		// Initialize redis client.
		storRedis = rediscli.NewRedis(&redisOptions)
		cache = rediscache.New(storRedis.Client(), db,
			rediscache.WithTimeout(settings.Common.LocalCacheTimeout, settings.Common.CacheTimeout),
			rediscache.WithVersion(settings.Common.CacheVersion),
		)

		// Initialize AMQP client and celery wrapper.
		amqpChannel, err = amqp.New(c.String("broker-url"), logg)
		if err != nil {
			return err
		}

		return nil
	}
	App.After = func(c *cli.Context) error {
		// Redis teardown.
		if storRedis != nil {
			storRedis.Shutdown() // nolint: errcheck
		}

		// Database teardown.
		if db != nil {
			db.Shutdown() // nolint: errcheck
		}

		// Sync loggers.
		if logger != nil {
			logger.Sync()
		}

		// Shutdown AMQP client.
		if amqpChannel != nil {
			amqpChannel.Shutdown()
		}

		// Shutdown job system.
		jobs.Shutdown()

		// Close tracing reporter.
		if jaegerExporter != nil {
			jaegerExporter.Flush()
		}

		// Flush remaining sentry events.
		sentry.Flush(5 * time.Second)

		return nil
	}
}
