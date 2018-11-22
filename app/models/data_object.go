package models

import (
	"fmt"
	"time"

	"github.com/go-pg/pg/orm"
)

// DataObject represents DataObject model.
type DataObject struct {
	State

	tableName struct{} `sql:"?schema.data_dataobject" pg:",discard_unknown_columns"` // nolint

	IsLive bool `sql:"_is_live"`

	ID        int
	Data      *Hstore `sql:"_data" state:"virtual"`
	Files     *Hstore `sql:"_files"`
	Revision  int
	CreatedAt *Time
	UpdatedAt *Time

	OwnerID int
	Owner   *User
	ClassID int    `sql:"_klass_id"`
	Class   *Class `pg:"fk:_klass_id"`
}

// NewDataObject ...
func NewDataObject(class *Class) *DataObject {
	return &DataObject{
		IsLive:    true,
		Class:     class,
		ClassID:   class.ID,
		Data:      NewHstore(),
		Files:     NewHstore(),
		Revision:  1,
		CreatedAt: NewTime(nil),
		UpdatedAt: NewTime(nil),
	}
}

func (m *DataObject) String() string {
	return fmt.Sprintf("DataObject<ID=%d, ClassID=%d>", m.ID, m.ClassID)
}

// VerboseName returns verbose name for model.
func (m *DataObject) VerboseName() string {
	return "Data Object"
}

// BeforeUpdate hook.
func (m *DataObject) BeforeUpdate(db orm.DB) error {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return nil
}
