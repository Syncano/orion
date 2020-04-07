package cmd

import (
	_ "expvar" // Register expvar default http handler.
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/go-redis/redis/v7"
	opentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
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
		&cli.StringFlag{
			Name: "dsn", Usage: "enable sentry logging",
			EnvVars: []string{"SENTRY_DSN"},
		},
		&cli.IntFlag{
			Name: "port, p", Usage: "port for expvar server",
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
			Name: "zipkin-addr", Usage: "zipkin address",
			EnvVars: []string{"ZIPKIN_ADDR"}, Value: "zipkin",
		},
		&cli.Float64Flag{
			Name: "tracing-sampling", Usage: "tracing sampling value",
			EnvVars: []string{"TRACING_SAMPLING"}, Value: 1,
		},
		&cli.StringFlag{
			Name: "service-name, n", Usage: "service name",
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
		if err := log.Init(c.String("dsn"), c.Bool("debug")); err != nil {
			return err
		}

		if err := raven.SetDSN(c.String("dsn")); err != nil {
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
		http.Handle("/metrics", promhttp.Handler())

		// Initialize tracing.
		reporter := zipkinhttp.NewReporter(fmt.Sprintf("http://%s:9411/api/v2/spans", c.String("zipkin-addr")))
		defer reporter.Close()

		endpoint, err := zipkin.NewEndpoint(c.String("service-name"), "")
		if err != nil {
			logger.With(zap.Error(err)).Fatal("Unable to create local endpoint error")
		}

		// Initialize tracer.
		nativeTracer, err := zipkin.NewTracer(reporter,
			zipkin.WithLocalEndpoint(endpoint),
			zipkin.WithSampler(zipkin.NewModuloSampler(uint64(1/c.Float64("tracing-sampling")))),
		)
		if err != nil {
			logger.With(zap.Error(err)).Fatal("Unable to create tracer")
		}

		// Use zipkin-go-opentracing to wrap our tracer.
		tracer := zipkinot.Wrap(nativeTracer)

		opentracing.SetGlobalTracer(tracer)

		// Initialize database client.
		storage.InitDB(dbOptions, dbInstancesOptions, c.Bool("debug"))

		// Initialize s3 client.
		storage.InitData()

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

		return nil
	}
}
