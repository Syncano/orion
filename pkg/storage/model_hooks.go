package storage

import (
	"fmt"
	"sync"

	"github.com/go-pg/pg/orm"
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

// AddModelDeleteHook ...
func AddModelDeleteHook(model interface{}, f DeleteModelHookFunc) {
	key := fmt.Sprintf("%T", model)
	modelhookmu.Lock()
	deleteHooks[key] = append(deleteHooks[key], f)
	modelhookmu.Unlock()
}

// AddModelSoftDeleteHook ...
func AddModelSoftDeleteHook(model interface{}, f SoftDeleteModelHookFunc) {
	key := fmt.Sprintf("%T", model)
	modelhookmu.Lock()
	softDeleteHooks[key] = append(softDeleteHooks[key], f)
	modelhookmu.Unlock()
}

// AddModelSaveHook ...
func AddModelSaveHook(model interface{}, f SaveModelHookFunc) {
	key := fmt.Sprintf("%T", model)
	modelhookmu.Lock()
	saveHooks[key] = append(saveHooks[key], f)
	modelhookmu.Unlock()
}

// ProcessModelSaveHook ...
func ProcessModelSaveHook(c DBContext, db orm.DB, created bool, model interface{}) error {
	key := fmt.Sprintf("%T", model)
	modelhookmu.RLock()
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

// ProcessModelDeleteHook ...
func ProcessModelDeleteHook(c DBContext, db orm.DB, model interface{}) error {
	key := fmt.Sprintf("%T", model)
	modelhookmu.RLock()
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

// ProcessModelSoftDeleteHook ...
func ProcessModelSoftDeleteHook(c DBContext, db orm.DB, model interface{}) error {
	key := fmt.Sprintf("%T", model)
	modelhookmu.RLock()
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
