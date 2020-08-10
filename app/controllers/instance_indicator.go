package controllers

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/v2/database"
	"github.com/Syncano/pkg-go/v2/database/manager"
)

func (ctr *Controller) updateInstanceIndicatorValue(c database.DBContext, db orm.DB, typ, diff int) error {
	instance := c.(echo.Context).Get(settings.ContextInstanceKey).(*models.Instance)
	mgr := ctr.q.NewInstanceIndicatorManager(c.(echo.Context))
	mgr.SetDB(db)

	o := &models.InstanceIndicator{InstanceID: instance.ID, Type: typ}
	if err := manager.Lock(mgr.ByInstanceAndType(o)); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	o.Value += diff

	return mgr.Update(o)
}
