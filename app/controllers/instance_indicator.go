package controllers

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

func (ctr *Controller) updateInstanceIndicatorValue(c echo.Context, db orm.DB, typ, diff int) error {
	instance := c.Get(settings.ContextInstanceKey).(*models.Instance)
	mgr := ctr.q.NewInstanceIndicatorManager(c)
	mgr.SetDB(db)

	o := &models.InstanceIndicator{InstanceID: instance.ID, Type: typ}
	if err := manager.Lock(mgr.ByInstanceAndTypeQ(o)); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	o.Value += diff

	return mgr.UpdateContext(c.Request().Context(), o)
}
