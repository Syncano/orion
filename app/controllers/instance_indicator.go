package controllers

import (
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/storage"
)

func (ctr *Controller) updateInstanceIndicatorValue(c storage.DBContext, db orm.DB, typ, diff int) error {
	instance := c.Get(settings.ContextInstanceKey).(*models.Instance)
	mgr := ctr.q.NewInstanceIndicatorManager(c)
	mgr.SetDB(db)

	o := &models.InstanceIndicator{InstanceID: instance.ID, Type: typ}
	if err := query.Lock(mgr.ByInstanceAndType(o)); err != nil {
		if err == pg.ErrNoRows {
			return api.NewNotFoundError(o)
		}

		return err
	}

	o.Value += diff

	return mgr.Update(o)
}
