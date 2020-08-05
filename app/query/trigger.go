package query

import (
	"fmt"

	"github.com/go-pg/pg/v9"
	json "github.com/json-iterator/go"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/pkg-go/v2/database"
	"github.com/Syncano/pkg-go/v2/database/fields"
	"github.com/Syncano/pkg-go/v2/database/manager"
	"github.com/Syncano/pkg-go/v2/util"
)

// TriggerManager represents Trigger manager.
type TriggerManager struct {
	*Factory
	*manager.Manager
}

// NewTriggerManager creates and returns new Trigger manager.
func (q *Factory) NewTriggerManager(c database.DBContext) *TriggerManager {
	return &TriggerManager{Factory: q, Manager: manager.NewTenantManager(q.db, c)}
}

// Match outputs one object within specific class filtered by id.
func (m *TriggerManager) Match(instance *models.Instance, event map[string]string, signal string) ([]*models.Trigger, error) {
	var o []*models.Trigger

	eventSerialized, e := json.ConfigCompatibleWithStandardLibrary.Marshal(event)
	util.Must(e)

	versionKey := fmt.Sprintf("i=%d;e=%s", instance.ID, eventSerialized)
	lookup := fmt.Sprintf("s=%s", signal)

	err := m.c.SimpleFuncCache("Trigger.Match", lookup, versionKey, o, func() (interface{}, error) {
		ehstore := new(fields.Hstore)
		ehstore.Set(event) // nolint: errcheck

		err := m.Query(&o).Where("event @> ?", ehstore).Where("signals @> ?", pg.Array([]string{signal})).Select()
		return o, err
	})

	return o, err
}
