package query

import (
	"fmt"

	"github.com/go-pg/pg"
	json "github.com/json-iterator/go"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/pkg/cache"
	"github.com/Syncano/orion/pkg/storage"
	"github.com/Syncano/orion/pkg/util"
)

// TriggerManager represents Trigger manager.
type TriggerManager struct {
	*Manager
}

// NewTriggerManager creates and returns new Trigger manager.
func NewTriggerManager(c storage.DBContext) *TriggerManager {
	return &TriggerManager{Manager: NewTenantManager(c)}
}

// Match outputs one object within specific class filtered by id.
func (mgr *TriggerManager) Match(instance *models.Instance, event map[string]string, signal string) ([]*models.Trigger, error) {
	var o []*models.Trigger

	eventSerialized, e := json.ConfigCompatibleWithStandardLibrary.Marshal(event)
	util.Must(e)

	versionKey := fmt.Sprintf("i=%d;e=%s", instance.ID, eventSerialized)
	lookup := fmt.Sprintf("s=%s", signal)

	err := cache.SimpleFuncCache("Trigger.Match", versionKey, o, lookup, func() (interface{}, error) {
		ehstore := new(models.Hstore)
		ehstore.Set(event) // nolint: errcheck

		err := mgr.Query(&o).Where("event @> ?", ehstore).Where("signals @> ?", pg.Array([]string{signal})).Select()
		return o, err
	})

	return o, err
}
