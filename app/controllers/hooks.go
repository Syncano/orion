package controllers

import (
	"reflect"
	"strings"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/tasks"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
)

func init() {
	// DataObject cleanup.
	storage.AddModelDeleteHook((*models.DataObject)(nil), dataObjectDeleteHook)

	// Triggers.
	storage.AddModelSoftDeleteHook((*models.DataObject)(nil), dataObjectSoftDeleteTriggerHook)

	// LiveObject cleanup.
	// TODO: InstanceIndicator post save hook after live obj delete is done.
	storage.AddModelSoftDeleteHook(storage.AnyModel, liveObjectSoftDeleteHook)

	// Cache invalidate hooks.
	for _, model := range []interface{}{
		(*models.AdminInstanceRole)(nil),
		(*models.Admin)(nil),
		(*models.APIKey)(nil),
		(*models.Profile)(nil),
		(*models.Subscription)(nil),
		(*models.Instance)(nil),
		(*models.Channel)(nil),
		(*models.Class)(nil),
		(*models.Codebox)(nil),
		(*models.SocketEndpoint)(nil),
		(*models.SocketEnvironment)(nil),
		(*models.Socket)(nil),
		(*models.User)(nil),
	} {
		storage.AddModelDeleteHook(model, cacheDeleteHook)
		storage.AddModelSoftDeleteHook(model, cacheDeleteHook)
		storage.AddModelSaveHook(model, cacheSaveHook)
	}
}

func cacheSaveHook(c storage.DBContext, db orm.DB, created bool, m interface{}) error {
	if created {
		return nil
	}

	cache.ModelCacheInvalidate(db, m)

	return nil
}

func cacheDeleteHook(c storage.DBContext, db orm.DB, m interface{}) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

func liveObjectSoftDeleteHook(c storage.DBContext, db orm.DB, m interface{}) error {
	table := orm.GetTable(reflect.TypeOf(m).Elem())
	n := strings.Split(string(table.FullName), ".")
	modelName := strings.ReplaceAll(n[len(n)-1], "_", ".")

	objectPK := table.PKs[0].Value(reflect.ValueOf(m).Elem()).Interface()

	storage.AddDBCommitHook(db, func() error {
		return tasks.NewDeleteLiveObjectTask(
			c.Get(settings.ContextInstanceKey).(*models.Instance).ID,
			modelName, objectPK,
		).Publish()
	})

	return nil
}
