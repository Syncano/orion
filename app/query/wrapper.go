package query

import (
	"github.com/labstack/echo/v4"

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
