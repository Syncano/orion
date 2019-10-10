package models

// StateField ...
//go:generate go run github.com/vektra/mockery/cmd/mockery -name StateField
type StateField interface {
	Get(o interface{}) interface{}
}

// OrderField ...
//go:generate go run github.com/vektra/mockery/cmd/mockery -name OrderField
type OrderField interface {
	SQLName() string
	Get(o interface{}) interface{}
	ToString(v interface{}) (string, error)
	FromString(s string) (interface{}, error)
}

// FilterField ...
//go:generate go run github.com/vektra/mockery/cmd/mockery -name FilterField
type FilterField interface {
	Name() string
	Type() string
	SQLName() string
}
