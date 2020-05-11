package storage

import (
	"net/http"

	"github.com/go-pg/pg/v9"
)

//go:generate go run github.com/vektra/mockery/cmd/mockery -name DBContext
type DBContext interface {
	// Get retrieves data from the context.
	Get(key string) interface{}

	// Set saves data in the context.
	Set(key string, val interface{})

	// Request returns `*http.Request`.
	Request() *http.Request
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name Databaser
type Databaser interface {
	DB() *pg.DB
	TenantDB(schema string) *pg.DB
}

var _ Databaser = (*Database)(nil)
