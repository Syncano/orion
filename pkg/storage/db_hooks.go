package storage

import (
	"fmt"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

type hookFunc func() error

func (d *Database) AddDBCommitHook(db orm.DB, f hookFunc) {
	tx, istx := db.(*pg.Tx)
	if !istx {
		_ = f()
		return
	}

	d.dbhookmu.Lock()
	d.commitHooks[tx] = append(d.commitHooks[tx], f)
	d.dbhookmu.Unlock()
}

func (d *Database) AddDBRollbackHook(db orm.DB, f hookFunc) {
	tx, istx := db.(*pg.Tx)
	if !istx {
		return
	}

	d.dbhookmu.Lock()
	d.rollbackHooks[tx] = append(d.rollbackHooks[tx], f)
	d.dbhookmu.Unlock()
}

func (d *Database) processDBHooks(tx *pg.Tx, process, hooks map[*pg.Tx][]hookFunc) error {
	d.dbhookmu.RLock()
	funcs := process[tx]
	_, hooksOK := hooks[tx]
	d.dbhookmu.RUnlock()

	if funcs != nil || hooksOK {
		d.dbhookmu.Lock()
		delete(process, tx)
		delete(hooks, tx)
		d.dbhookmu.Unlock()
	}

	for _, f := range funcs {
		if err := f(); err != nil {
			return err
		}
	}

	return nil
}

// ProcessDBCommitHooks processes all commit hooks for transaction and removes rollback hooks if any.
// Needs to be called after every transaction.
func (d *Database) ProcessDBCommitHooks(tx *pg.Tx) error {
	return d.processDBHooks(tx, d.commitHooks, d.rollbackHooks)
}

// ProcessDBRollbackHooks processes all rollback hooks for transaction and removes commit hooks if any.
// Needs to be called after every transaction.
func (d *Database) ProcessDBRollbackHooks(tx *pg.Tx) error {
	return d.processDBHooks(tx, d.rollbackHooks, d.commitHooks)
}

// RunTransactionWithHooks is a helper method that calls commit and rollback hooks.
func (d *Database) RunTransactionWithHooks(tx *pg.Tx, fn func(*pg.Tx) error) error {
	defer func() {
		if err := recover(); err != nil {
			_ = tx.Rollback()

			panic(err)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		err = fmt.Errorf("rollback due to error: %w", err)

		if hookErr := d.ProcessDBRollbackHooks(tx); hookErr != nil {
			return fmt.Errorf("hook error: %w", hookErr)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return d.ProcessDBCommitHooks(tx)
}
