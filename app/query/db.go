package query

import (
	"github.com/go-pg/pg"

	"github.com/Syncano/orion/pkg/storage"
)

const ContextSchemaKey = "schema"

// DB returns base db for context.
func DB(c storage.DBContext) *pg.DB {
	return storage.DB()
}

// TenantDB returns base tenant db for context.
func TenantDB(c storage.DBContext) *pg.DB {
	schema := c.Get(ContextSchemaKey).(string)
	return storage.TenantDB(schema)
}
