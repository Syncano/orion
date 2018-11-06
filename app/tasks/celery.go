package tasks

import (
	"github.com/Syncano/orion/pkg/celery"
)

const (
	codeboxQueue                 = "codebox"
	celeryHandleTriggerEventTask = "apps.triggers.tasks.HandleTriggerEventTask"

	defaultQueue               = "default"
	celeryDeleteLiveObjectTask = "apps.core.tasks.DeleteLiveObjectTask"
)

// NewCeleryHandleTriggerEventTask returns a new handle trigger task object.
func NewCeleryHandleTriggerEventTask(instancePK int, event map[string]string, signal string,
	data map[string]interface{},
	kwargs map[string]interface{}) *celery.Task {

	kw := map[string]interface{}{
		"instance_pk": instancePK,
	}
	for k, v := range kwargs {
		kw[k] = v
	}
	return celery.NewTask(celeryHandleTriggerEventTask, codeboxQueue, []interface{}{
		event,
		signal,
		data,
	}, kw)
}

// NewDeleteLiveObjectTask returns a new handle trigger task object.
func NewDeleteLiveObjectTask(instancePK int, modelName string, objectPK interface{}) *celery.Task {
	return celery.NewTask(celeryDeleteLiveObjectTask, defaultQueue, []interface{}{
		modelName,
		objectPK,
	}, map[string]interface{}{
		"instance_pk": instancePK,
	})
}
