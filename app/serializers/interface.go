package serializers

// Serializer ...
//go:generate mockery -name Serializer
type Serializer interface {
	Response(interface{}) interface{}
}
