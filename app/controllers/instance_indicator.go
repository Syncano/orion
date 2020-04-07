package controllers

import (
	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
)

func updateInstanceIndicatorValue(c storage.DBContext, db orm.DB, typ, diff int) error {
	instance := c.Get(settings.ContextInstanceKey).(*models.Instance)
	mgr := query.NewInstanceIndicatorManager(c)
	mgr.SetDB(db)

	o := &models.InstanceIndicator{InstanceID: instance.ID, Type: typ}
	if query.Lock(mgr.ByInstanceAndType(o)) != nil {
		return api.NewNotFoundError(o)
	}

	o.Value += diff

	return mgr.Update(o)
}
