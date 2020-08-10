package api

import (
	"context"

	"github.com/go-pg/pg/v9"
)

//go:generate go run github.com/vektra/mockery/cmd/mockery -inpkg -testonly -name Verboser
type Verboser interface {
	VerboseName() string
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -inpkg -testonly -name Deleter
type Deleter interface {
	DeleteContext(context.Context, interface{}) error
	RunInTransaction(func(tx *pg.Tx) error) error
}
