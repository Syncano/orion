package models

//go:generate go run github.com/vektra/mockery/cmd/mockery -name StateField
type StateField interface {
	Get(o interface{}) interface{}
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name OrderField
type OrderField interface {
	SQLName() string
	Get(o interface{}) interface{}
	ToString(v interface{}) (string, error)
	FromString(s string) (interface{}, error)
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name FilterField
type FilterField interface {
	Name() string
	Type() string
	SQLName() string
}
