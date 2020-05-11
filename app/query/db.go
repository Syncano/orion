package query

import (
	"github.com/go-pg/pg/v9"

	"github.com/Syncano/orion/pkg/storage"
)

const ContextSchemaKey = "schema"

// DB returns base db for context.
func DB(db storage.Databaser, c storage.DBContext) *pg.DB {
	return db.DB()
}

// TenantDB returns base tenant db for context.
func TenantDB(db storage.Databaser, c storage.DBContext) *pg.DB {
	schema := c.Get(ContextSchemaKey).(string)
	return db.TenantDB(schema)
}
