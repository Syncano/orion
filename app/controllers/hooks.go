package controllers

import (
	"reflect"
	"strings"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/tasks"
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
}

func liveObjectSoftDeleteHook(c storage.DBContext, db orm.DB, m interface{}) error {
	table := orm.GetTable(reflect.TypeOf(m).Elem())
	n := strings.Split(string(table.Name), ".")
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
