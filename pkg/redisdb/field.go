package redisdb

import (
	"reflect"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/Syncano/orion/pkg/settings"
)

// FieldAdapter is an interface for Redis model field adapter.
type FieldAdapter interface {
	Load(value string) interface{}
	Dump(value interface{}) string
}

var (
	fieldAdapterType = reflect.TypeOf((*FieldAdapter)(nil)).Elem()
	timeType         = reflect.TypeOf((*time.Time)(nil)).Elem()
	jsonAdapter      = &jsonFieldAdapter{}
	timeAdapter      = &datetimeFieldAdapter{}
)

// Field represents a Redis model field.
type Field struct {
	Name    string
	Field   reflect.StructField
	Type    reflect.Type
	Adapter FieldAdapter

	def string
}

// Value returns field value for given struct value.
func (f *Field) Value(v reflect.Value) reflect.Value {
	return v.FieldByIndex(f.Field.Index)
}

// Default returns default value for that field.
func (f *Field) Default() reflect.Value {
	return reflect.ValueOf(f.Adapter.Load(f.def))
}

// HasDefault returns true if default is specified.
func (f *Field) HasDefault() bool {
	return f.def != ""
}

var adaptersMap = map[reflect.Kind]FieldAdapter{
	reflect.String: &stringFieldAdapter{},
	reflect.Int:    &integerFieldAdapter{},
	reflect.Map:    jsonAdapter,
	reflect.Slice:  jsonAdapter,
	reflect.Bool:   &boolFieldAdapter{},
}

func getAdapter(sf reflect.StructField) FieldAdapter {
	if sf.Type == timeType {
		return timeAdapter
	}
	if sf.Type.Implements(fieldAdapterType) {
		return reflect.New(sf.Type).Interface().(FieldAdapter)
	}

	kind := sf.Type.Kind()
	return adaptersMap[kind]
}

// String
type stringFieldAdapter struct{}

func (f *stringFieldAdapter) Load(value string) interface{} {
	return value
}

func (f *stringFieldAdapter) Dump(value interface{}) string {
	return value.(string)
}

// Integer
type integerFieldAdapter struct{}

func (f *integerFieldAdapter) Load(value string) interface{} {
	v, _ := strconv.Atoi(value)
	return v
}

func (f *integerFieldAdapter) Dump(value interface{}) string {
	return strconv.Itoa(value.(int))
}

// Datetime
type datetimeFieldAdapter struct{}

func (f *datetimeFieldAdapter) Load(value string) interface{} {
	t, _ := time.Parse(settings.Common.DateTimeFormat, value)
	return t
}

func (f *datetimeFieldAdapter) Dump(value interface{}) string {
	return value.(time.Time).Format(settings.Common.DateTimeFormat)
}

// JSON
type jsonFieldAdapter struct{}

func (f *jsonFieldAdapter) Load(value string) interface{} {
	var v interface{}
	jsoniter.Unmarshal([]byte(value), &v) // nolint: errcheck
	return v
}

func (f *jsonFieldAdapter) Dump(value interface{}) string {
	if b, err := jsoniter.Marshal(value); err == nil {
		return string(b)
	}
	return ""
}

// Bool
type boolFieldAdapter struct{}

func (f *boolFieldAdapter) Load(value string) interface{} {
	return value == "t"
}

func (f *boolFieldAdapter) Dump(value interface{}) string {
	if value.(bool) {
		return "t"
	}
	return "f"
}
