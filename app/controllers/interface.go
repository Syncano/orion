package controllers

import (
	"github.com/go-pg/pg/orm"
	"github.com/labstack/echo"
)

// Paginator ...
//go:generate mockery -inpkg -testonly -name Paginator
type Paginator interface {
	FilterObjects(cursor Cursorer) (*orm.Query, error)
	CreateCursor(c echo.Context, defaultOrderAsc bool) Cursorer
}

// Assert interface compatibility.
var (
	_ Paginator = (*PaginatorDB)(nil)
	_ Paginator = (*PaginatorOrderedDB)(nil)
)

// Cursorer ...
//go:generate mockery -inpkg -testonly -name Cursorer
type Cursorer interface {
	NextURL(path string) string
	PrevURL(path string) string

	Limit() int
	LastPK() int
	IsOrderAsc() bool
	IsForward() bool
	SetFirst(interface{})
	SetLast(interface{})
}

// Assert interface compatibility.
var (
	_ Cursorer = (*cursor)(nil)
	_ Cursorer = (*keysetcursor)(nil)
)
