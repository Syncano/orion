package query

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/v2/database"
)

type wrappedContext struct {
	echo.Context
	schemaGetter func() string
}

func (w *wrappedContext) Schema() string {
	return w.schemaGetter()
}

func WrapContext(c echo.Context) database.DBContext {
	return &wrappedContext{
		Context: c,
		schemaGetter: func() string {
			return c.Get(settings.ContextSchemaKey).(string)
		},
	}
}

type DBContexter interface {
	DBContext() database.DBContext
}

func DBToStdContext(m DBContexter) context.Context {
	return m.DBContext().(echo.Context).Request().Context()
}

func TenantDB(c echo.Context, db *database.DB) *pg.DB {
	return db.TenantDB(c.Get(settings.ContextSchemaKey).(string))
}

type Inserter interface {
	InsertContext(ctx context.Context, o interface{}) error
}

func ValidateAndInsert(c echo.Context, mgr Inserter, validator, obj interface{}, bind func()) error {
	return api.BindValidateAndExec(c, validator, func() error {
		bind()

		return mgr.InsertContext(c.Request().Context(), obj)
	})
}
