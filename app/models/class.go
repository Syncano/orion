package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/mitchellh/mapstructure"

	"github.com/Syncano/pkg-go/database/fields"
)

const UserClassName = "user_profile"

// Class represents Class model.
type Class struct {
	State
	tableName struct{} `pg:"?schema.data_klass,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

	ID              int
	Name            string
	Revision        int
	Schema          fields.JSON
	Mapping         fields.Hstore
	ExistingIndexes fields.JSON
	IndexChanges    fields.JSON
	Refs            fields.JSON
	Visible         bool
	CreatedAt       fields.Time
	UpdatedAt       fields.Time
	Metadata        fields.JSON
	Description     string

	computedSchema map[string]*DataObjectField
	ObjectsCount   int           `pg:"-" msgpack:"-"`
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
func (m *Class) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return ctx, nil
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
