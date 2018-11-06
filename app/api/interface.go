package api

import "github.com/go-pg/pg"

// Verboser ...
//go:generate mockery -inpkg -testonly -name Verboser
type Verboser interface {
	VerboseName() string
}

// Deleter ...
//go:generate mockery -inpkg -testonly -name Deleter
type Deleter interface {
	Delete(interface{}) error
	RunInTransaction(func(tx *pg.Tx) error) error
}
