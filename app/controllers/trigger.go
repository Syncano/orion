package controllers

import (
	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
	"github.com/Syncano/orion/pkg/util"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/tasks"
)

func dataObjectSoftDeleteTriggerHook(c storage.DBContext, db orm.DB, i interface{}) error {
	launchDataObjectTrigger(c, db, i.(*models.DataObject), models.TriggerSignalDelete)
	return nil
}

func launchDataObjectTrigger(c storage.DBContext, db orm.DB, o *models.DataObject, signal string) {
	var (
		changes []string
	)
	class := c.Get(contextClassKey).(*models.Class)
	if signal == models.TriggerSignalUpdate {
		changes = o.SQLChangesVirtual()
	}
	launchTrigger(c, db, o, map[string]string{"source": "dataobject", "class": class.Name}, signal, serializers.DataObjectSerializer{Class: class}, changes)
}

func launchTrigger(c storage.DBContext, db orm.DB, o interface{}, event map[string]string, signal string, serializer serializers.Serializer, changes []string) {
	var (
		data map[string]interface{}
	)
	instance := c.Get(settings.ContextInstanceKey).(*models.Instance)
	if t, _ := query.NewTriggerManager(c).Match(instance, event, signal); len(t) == 0 {
		return
	}

	if signal == models.TriggerSignalCreate || signal == models.TriggerSignalUpdate {
		data = serializer.Response(o).(map[string]interface{})
	}

	// Filter out unchanged fields on update.
	if signal == models.TriggerSignalUpdate && len(changes) > 0 {
		dtemp := make(map[string]interface{})

		for _, change := range changes {
			dtemp[change] = data[change]
		}
		data = dtemp
	}

	storage.AddDBCommitHook(db, func() {
		util.Must(
			tasks.NewCeleryHandleTriggerEventTask(
				instance.ID,
				event, signal, data, map[string]interface{}{"changes": changes},
			).Publish(),
		)
	})
}
