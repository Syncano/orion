package redisdb

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/Syncano/orion/pkg/util"
)

const pkName = "id"

func indirectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// Table represents a Redis model table created from Go struct.
type Table struct {
	Type reflect.Type

	Name    string
	pkField *Field

	Fields map[string]*Field
}

func newTable(typ reflect.Type) *Table {
	r := new(Table)
	r.Type = typ
	r.Name = typ.Name()

	r.Fields = make(map[string]*Field, r.Type.NumField())
	r.addFields(r.Type, nil)

	return r
}

// PK returns PK of table.
func (r *Table) PK(v reflect.Value) int {
	return int(r.pkField.Value(v).Int())
}

// SetPK sets PK of table.
func (r *Table) SetPK(v reflect.Value, pk int) {
	r.pkField.Value(v).SetInt(int64(pk))
}

func (r *Table) addFields(typ reflect.Type, baseIndex []int) {
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		// Make a copy so slice is not shared between fields.
		index := make([]int, len(baseIndex))
		copy(index, baseIndex)

		if f.Anonymous {
			if tag := f.Tag.Get("redis"); tag == "-" {
				continue
			}

			fieldType := indirectType(f.Type)
			r.addFields(fieldType, append(index, f.Index...))
			continue
		}

		field := r.newField(typ, f)
		if field != nil {
			r.Fields[field.Name] = field
		}
	}

	if _, ok := r.Fields[pkName]; !ok || r.Fields[pkName].Field.Type.Kind() != reflect.Int {
		panic("redis: pk undefined on model or wrong type (int required)")
	}
	r.pkField = r.Fields[pkName]
}

func (r *Table) newField(typ reflect.Type, f reflect.StructField) *Field {
	name := f.Name

	if tag := f.Tag.Get("redis"); tag != "" {
		if tag == "-" {
			return nil
		}
		name = tag
	} else {
		name = util.Underscore(name)
	}

	adapter := getAdapter(f)
	if adapter == nil {
		return nil
	}

	return &Field{
		Name:    name,
		Field:   f,
		Type:    typ,
		Adapter: adapter,
		def:     f.Tag.Get("default"),
	}
}

var _tables = &tables{
	tables: make(map[reflect.Type]*Table),
}

type tables struct {
	mu     sync.RWMutex
	tables map[reflect.Type]*Table
}

// GetTable returns table for specified type.
func GetTable(typ reflect.Type) *Table {
	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("redis: got %s, wanted %s", typ.Kind(), reflect.Struct))
	}

	_tables.mu.RLock()
	table, ok := _tables.tables[typ]
	_tables.mu.RUnlock()
	if ok {
		return table
	}

	_tables.mu.Lock()
	table, ok = _tables.tables[typ]
	if ok {
		_tables.mu.Unlock()
		return table
	}

	table = newTable(typ)
	_tables.tables[typ] = table
	_tables.mu.Unlock()
	return table
}
