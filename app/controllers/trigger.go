package controllers

import (
	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/app/tasks"
	"github.com/Syncano/pkg-go/database"
)

func (ctr *Controller) dataObjectSoftDeleteTriggerHook(c database.DBContext, db orm.DB, i interface{}) error {
	ctr.launchDataObjectTrigger(c, db, i.(*models.DataObject), models.TriggerSignalDelete)
	return nil
}

func (ctr *Controller) launchDataObjectTrigger(c database.DBContext, db orm.DB, o *models.DataObject, signal string) {
	var (
		changes []string
	)

	class := c.Get(contextClassKey).(*models.Class)

	if signal == models.TriggerSignalUpdate {
		changes = o.SQLChangesVirtual()
	}

	ctr.launchTrigger(c, db, o, map[string]string{"source": "dataobject", "class": class.Name}, signal, serializers.DataObjectSerializer{Class: class}, changes)
}

func (ctr *Controller) launchTrigger(c database.DBContext, db orm.DB, o interface{}, event map[string]string, signal string, serializer serializers.Serializer, changes []string) {
	var (
		data map[string]interface{}
	)

	instance := c.Get(settings.ContextInstanceKey).(*models.Instance)
	if t, _ := ctr.q.NewTriggerManager(c).Match(instance, event, signal); len(t) == 0 {
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

	ctr.db.AddDBCommitHook(db, func() error {
		return tasks.NewCeleryHandleTriggerEventTask(
			instance.ID,
			event, signal, data, map[string]interface{}{"changes": changes},
		).Publish(ctr.cel)
	})
}
