package storage

import (
	"fmt"
	"sync"

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
	modelhookmu     sync.RWMutex
	saveHooks       = map[string][]SaveModelHookFunc{}
	deleteHooks     = map[string][]DeleteModelHookFunc{}
	softDeleteHooks = map[string][]SoftDeleteModelHookFunc{}

	// AnyModel signifies any model hook.
	AnyModel    = (*anyModel)(nil)
	anyModelKey = fmt.Sprintf("%T", AnyModel)
)

func AddModelDeleteHook(model interface{}, f DeleteModelHookFunc) {
	modelhookmu.Lock()

	key := fmt.Sprintf("%T", model)
	deleteHooks[key] = append(deleteHooks[key], f)

	modelhookmu.Unlock()
}

func AddModelSoftDeleteHook(model interface{}, f SoftDeleteModelHookFunc) {
	modelhookmu.Lock()

	key := fmt.Sprintf("%T", model)
	softDeleteHooks[key] = append(softDeleteHooks[key], f)

	modelhookmu.Unlock()
}

func AddModelSaveHook(model interface{}, f SaveModelHookFunc) {
	modelhookmu.Lock()

	key := fmt.Sprintf("%T", model)
	saveHooks[key] = append(saveHooks[key], f)

	modelhookmu.Unlock()
}

func ProcessModelSaveHook(c DBContext, db orm.DB, created bool, model interface{}) error {
	modelhookmu.RLock()

	key := fmt.Sprintf("%T", model)
	funcs := saveHooks[key]
	funcsAny := saveHooks[anyModelKey]

	modelhookmu.RUnlock()

	funcs = append(funcs, funcsAny...)
	for _, f := range funcs {
		if err := f(c, db, created, model); err != nil {
			return err
		}
	}

	return nil
}

func ProcessModelDeleteHook(c DBContext, db orm.DB, model interface{}) error {
	modelhookmu.RLock()

	key := fmt.Sprintf("%T", model)
	funcs := deleteHooks[key]
	funcsAny := deleteHooks[anyModelKey]

	modelhookmu.RUnlock()

	funcs = append(funcs, funcsAny...)
	for _, f := range funcs {
		if err := f(c, db, model); err != nil {
			return err
		}
	}

	return nil
}

func ProcessModelSoftDeleteHook(c DBContext, db orm.DB, model interface{}) error {
	modelhookmu.RLock()

	key := fmt.Sprintf("%T", model)
	funcs := softDeleteHooks[key]
	funcsAny := softDeleteHooks[anyModelKey]

	modelhookmu.RUnlock()

	funcs = append(funcs, funcsAny...)
	for _, f := range funcs {
		if err := f(c, db, model); err != nil {
			return err
		}
	}

	return nil
}
