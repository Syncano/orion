package serializers

//go:generate go run github.com/vektra/mockery/cmd/mockery -name Serializer
type Serializer interface {
	Response(interface{}) interface{}
}
