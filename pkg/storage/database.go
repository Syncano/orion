package storage

import (
	"context"
	"net"
	"time"

	"github.com/go-pg/pg/v9"
	"go.uber.org/zap"

	"github.com/Syncano/orion/pkg/log"
	"github.com/Syncano/orion/pkg/util"
)

type key int

const (
	// KeySchema is used in Context as a key to describe Schema.
	KeySchema        key = iota
	dbConnRetries        = 10
	dbConnRetrySleep     = 250 * time.Millisecond
)

var (
	commonDB *pg.DB
	tenantDB *pg.DB
)

// DefaultDBOptions returns
func DefaultDBOptions() *pg.Options {
	return &pg.Options{
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var conn net.Conn

			return conn, util.Retry(dbConnRetries, dbConnRetrySleep, func() error {
				var (
					err error
				)
				d := net.Dialer{Timeout: 3 * time.Second}
				conn, err = d.DialContext(ctx, network, addr)
				return err
			})
		},
		PoolSize:    10,
		IdleTimeout: 5 * time.Minute,
		PoolTimeout: 30 * time.Second,
		MaxConnAge:  15 * time.Minute,
		MaxRetries:  1,
	}
}

// InitDB sets up a database.
func InitDB(opts, instancesOpts *pg.Options, debug bool) {
	commonDB = initDB(opts, debug)
	tenantDB = commonDB

	if instancesOpts.Addr != opts.Addr || instancesOpts.Database != opts.Database {
		tenantDB = initDB(instancesOpts, debug)
	}
}

type debugHook struct {
	logger *zap.Logger
}

func (*debugHook) BeforeQuery(ctx context.Context, ev *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (h *debugHook) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	query, err := event.FormattedQuery()
	if err != nil {
		panic(err)
	}

	h.logger.Debug("Query",
		zap.String("query", query),
		zap.Duration("took", time.Since(event.StartTime)),
	)

	return nil
}

func initDB(opts *pg.Options, debug bool) *pg.DB {
	db := pg.Connect(opts)

	if debug {
		db.AddQueryHook(&debugHook{logger: log.Logger().WithOptions(zap.AddCallerSkip(8))})
	}

	return db
}

// DB returns database client.
func DB() *pg.DB {
	return commonDB
}

// TenantDB returns database client.
func TenantDB(schema string) *pg.DB {
	return tenantDB.WithParam("schema", pg.Ident(schema)).WithContext(context.WithValue(context.Background(), KeySchema, schema))
}
