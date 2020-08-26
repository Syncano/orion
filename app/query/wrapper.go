package query

import (
	"github.com/go-pg/pg/v9"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/v2/database"
)

func WrapContext(c echo.Context) database.DBContext {
	return database.WrapContextWithSchemaGetter(c.Request().Context(), func() string {
		schema := c.Get(settings.ContextSchemaKey)
		if schema == nil {
			return "public"
		}

		return schema.(string)
	}, c)
}

func TenantDB(c echo.Context, db *database.DB) *pg.DB {
	return db.TenantDB(c.Get(settings.ContextSchemaKey).(string))
}

type Inserter interface {
	Insert(o interface{}) error
}

func ValidateAndInsert(c echo.Context, mgr Inserter, validator, obj interface{}, bind func()) error {
	return api.BindValidateAndExec(c, validator, func() error {
		bind()

		return mgr.Insert(obj)
	})
}
