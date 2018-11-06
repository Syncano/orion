package storage

import (
	"fmt"
	"sync"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type hookFunc func()

var (
	dbhookmu      sync.RWMutex
	commitHooks   = map[*pg.Tx][]hookFunc{}
	rollbackHooks = map[*pg.Tx][]hookFunc{}
)

// AddDBCommitHook ...
func AddDBCommitHook(db orm.DB, f hookFunc) {
	tx, istx := db.(*pg.Tx)
	if !istx {
		f()
		return
	}
	dbhookmu.Lock()
	commitHooks[tx] = append(commitHooks[tx], f)
	dbhookmu.Unlock()
}

// AddDBRollbackHook ...
func AddDBRollbackHook(db orm.DB, f hookFunc) {
	tx, istx := db.(*pg.Tx)
	if !istx {
		return
	}
	dbhookmu.Lock()
	rollbackHooks[tx] = append(rollbackHooks[tx], f)
	dbhookmu.Unlock()
}

func processDBHooks(tx *pg.Tx, process map[*pg.Tx][]hookFunc, cleanup map[*pg.Tx][]hookFunc) error {
	dbhookmu.RLock()
	funcs := process[tx]
	_, cleanupNeeded := cleanup[tx]
	dbhookmu.RUnlock()

	if funcs != nil || cleanupNeeded {
		dbhookmu.Lock()
		delete(process, tx)
		delete(cleanup, tx)
		dbhookmu.Unlock()
	}

	for _, f := range funcs {
		f()
	}
	return nil
}

// ProcessDBCommitHooks processes all commit hooks for transaction and removes rollback hooks if any.
// Needs to be called after every transaction.
func ProcessDBCommitHooks(tx *pg.Tx) error {
	return processDBHooks(tx, commitHooks, rollbackHooks)
}

// ProcessDBRollbackHooks processes all rollback hooks for transaction and removes commit hooks if any.
// Needs to be called after every transaction.
func ProcessDBRollbackHooks(tx *pg.Tx) error {
	return processDBHooks(tx, rollbackHooks, commitHooks)
}

// RunTransactionWithHooks is a helper method that calls commit and rollback hooks.
func RunTransactionWithHooks(tx *pg.Tx, fn func(*pg.Tx) error) error {
	defer func() {
		if err := recover(); err != nil {
			_ = tx.Rollback()
			panic(err)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		if hookErr := ProcessDBRollbackHooks(tx); hookErr != nil {
			return fmt.Errorf("rollback error: %s, hook error: %s", err, hookErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return ProcessDBCommitHooks(tx)

}
