package storage

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-pg/pg"
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
		Dialer: func(network, addr string) (net.Conn, error) {
			var conn net.Conn
			return conn, util.Retry(dbConnRetries, dbConnRetrySleep, func() error {
				var err error
				conn, err = net.DialTimeout(network, addr, 3*time.Second)
				return err
			})
		},
		PoolSize:    10,
		IdleTimeout: 5 * time.Minute,
		PoolTimeout: 30 * time.Second,
		MaxConnAge:  15 * time.Minute,
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

func initDB(opts *pg.Options, debug bool) *pg.DB {
	db := pg.Connect(opts)

	if debug {
		logger := log.Logger()

		db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
			query, err := event.FormattedQuery()
			if err != nil {
				panic(err)
			}

			logger.With(
				zap.String("query", query),
				zap.Duration("took", time.Since(event.StartTime)),
			).Debug(fmt.Sprintf("%s:%d", event.File, event.Line))
		})
	}

	return db
}

// DB returns database client.
func DB() *pg.DB {
	return commonDB
}

// TenantDB returns database client.
func TenantDB(schema string) *pg.DB {
	return tenantDB.WithParam("schema", pg.F(schema)).WithContext(context.WithValue(context.Background(), KeySchema, schema))
}
