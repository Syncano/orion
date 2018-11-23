package models

import (
	"fmt"
	"time"

	"github.com/go-pg/pg/orm"
	"github.com/jackc/pgx/pgtype"
	"github.com/mitchellh/mapstructure"

	"github.com/Syncano/orion/pkg/cache"
)

// UserClassName ...
const UserClassName = "user_profile"

// Class represents Class model.
type Class struct {
	State
	tableName struct{} `sql:"?schema.data_klass" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID              int
	Name            string
	Revision        int
	Schema          JSON
	Mapping         Hstore
	ExistingIndexes JSON
	IndexChanges    JSON
	Refs            JSON
	Visible         bool
	CreatedAt       Time
	UpdatedAt       Time
	Metadata        JSON
	Description     string

	computedSchema map[string]*DataObjectField
	ObjectsCount   int           `sql:"-" msgpack:"-"`
	Objects        []*DataObject `pg:"fk:_klass_id" msgpack:"-"`
}

func (m *Class) String() string {
	return fmt.Sprintf("Class<ID=%d Name=%q>", m.ID, m.Name)
}

// VerboseName returns verbose name for model.
func (m *Class) VerboseName() string {
	return "Class"
}

// BeforeUpdate hook.
func (m *Class) BeforeUpdate(db orm.DB) error {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return nil
}

// AfterUpdate hook.
func (m *Class) AfterUpdate(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// AfterDelete hook.
func (m *Class) AfterDelete(db orm.DB) error {
	cache.ModelCacheInvalidate(db, m)
	return nil
}

// IsLocked returns true if object is locked.
func (m *Class) IsLocked() bool {
	return !m.IndexChanges.IsNull()
}

// GetStatus returns verbose name of current status.
func (m *Class) GetStatus() string {
	if m.IsLocked() {
		return "migrating"
	}
	return "ready"
}

// ComputedSchema computes and returns schema map.
func (m *Class) ComputedSchema() map[string]*DataObjectField {
	tableAlias := "data_object"
	if m.Name == UserClassName {
		tableAlias = "profile"
	}

	if m.computedSchema == nil {
		ret := make(map[string]*DataObjectField)

		for _, f := range m.Schema.Get().([]interface{}) {
			m := f.(map[string]interface{})
			field := &DataObjectField{TableAlias: tableAlias}
			if mapstructure.Decode(m, field) == nil {
				ret[field.FName] = field
			}
		}

		if m := m.Mapping.Get(); m != nil {
			mapp := m.(map[string]pgtype.Text)
			for k, v := range mapp {
				if f, ok := ret[k]; ok {
					f.Mapping = v.String
				}
			}
		}

		m.computedSchema = ret
	}
	return m.computedSchema
}

// FilterFields ...
func (m *Class) FilterFields() map[string]FilterField {
	filterFields := make(map[string]FilterField)
	def := defaultObjectFilterFields
	if m.Name == UserClassName {
		def = defaultUserFilterFields
	}

	for name, field := range def {
		filterFields[name] = field
	}
	for name, field := range m.ComputedSchema() {
		if field.FilterIndex {
			filterFields[name] = field
		}
	}
	return filterFields
}

// OrderFields ...
func (m *Class) OrderFields() map[string]OrderField {
	orderFields := make(map[string]OrderField)
	def := defaultObjectOrderFields
	if m.Name == UserClassName {
		def = defaultUserOrderFields
	}

	for name, field := range def {
		orderFields[name] = field
	}
	for name, field := range m.ComputedSchema() {
		if field.OrderIndex {
			orderFields[name] = field
		}
	}
	return orderFields
}
