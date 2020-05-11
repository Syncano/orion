package storage

import (
	"fmt"

	"github.com/go-pg/pg/v9/orm"
)

type (
	// SaveModelHookFunc is save model hook function definition.
	SaveModelHookFunc func(c DBContext, db orm.DB, created bool, m interface{}) error
	// SoftDeleteModelHookFunc is soft delete model hook function definition.
	SoftDeleteModelHookFunc func(c DBContext, db orm.DB, m interface{}) error
	// DeleteModelHookFunc is delete model hook function definition.
	DeleteModelHookFunc func(c DBContext, db orm.DB, m interface{}) error
)
type anyModel struct{}

var (
	// AnyModel signifies any model hook.
	AnyModel    = (*anyModel)(nil)
	anyModelKey = fmt.Sprintf("%T", AnyModel)
)

func (d *Database) AddModelDeleteHook(model interface{}, f DeleteModelHookFunc) {
	d.modelhookmu.Lock()

	key := fmt.Sprintf("%T", model)
	d.deleteHooks[key] = append(d.deleteHooks[key], f)

	d.modelhookmu.Unlock()
}

func (d *Database) AddModelSoftDeleteHook(model interface{}, f SoftDeleteModelHookFunc) {
	d.modelhookmu.Lock()

	key := fmt.Sprintf("%T", model)
	d.softDeleteHooks[key] = append(d.softDeleteHooks[key], f)

	d.modelhookmu.Unlock()
}

func (d *Database) AddModelSaveHook(model interface{}, f SaveModelHookFunc) {
	d.modelhookmu.Lock()

	key := fmt.Sprintf("%T", model)
	d.saveHooks[key] = append(d.saveHooks[key], f)

	d.modelhookmu.Unlock()
}

func (d *Database) ProcessModelSaveHook(c DBContext, db orm.DB, created bool, model interface{}) error {
	d.modelhookmu.RLock()

	key := fmt.Sprintf("%T", model)
	funcs := d.saveHooks[key]
	funcsAny := d.saveHooks[anyModelKey]

	d.modelhookmu.RUnlock()

	funcs = append(funcs, funcsAny...)
	for _, f := range funcs {
		if err := f(c, db, created, model); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) ProcessModelDeleteHook(c DBContext, db orm.DB, model interface{}) error {
	d.modelhookmu.RLock()

	key := fmt.Sprintf("%T", model)
	funcs := d.deleteHooks[key]
	funcsAny := d.deleteHooks[anyModelKey]

	d.modelhookmu.RUnlock()

	funcs = append(funcs, funcsAny...)
	for _, f := range funcs {
		if err := f(c, db, model); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) ProcessModelSoftDeleteHook(c DBContext, db orm.DB, model interface{}) error {
	d.modelhookmu.RLock()

	key := fmt.Sprintf("%T", model)
	funcs := d.softDeleteHooks[key]
	funcsAny := d.softDeleteHooks[anyModelKey]

	d.modelhookmu.RUnlock()

	funcs = append(funcs, funcsAny...)
	for _, f := range funcs {
		if err := f(c, db, model); err != nil {
			return err
		}
	}

	return nil
}
