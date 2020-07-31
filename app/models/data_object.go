package models

import (
	"context"
	"fmt"
	"time"

	"github.com/Syncano/pkg-go/v2/database/fields"
)

// DataObject represents DataObject model.
type DataObject struct {
	State

	tableName struct{} `pg:"?schema.data_dataobject,discard_unknown_columns"` // nolint

	IsLive bool `pg:"_is_live"`

	ID        int
	Data      fields.Hstore `pg:"_data" state:"virtual"`
	Files     fields.Hstore `pg:"_files"`
	Revision  int
	CreatedAt fields.Time
	UpdatedAt fields.Time

	OwnerID int
	Owner   *User
	ClassID int    `pg:"_klass_id"`
	Class   *Class `pg:"fk:_klass_id"`
}

func NewDataObject(class *Class) *DataObject {
	now := time.Now()

	return &DataObject{
		IsLive:    true,
		Class:     class,
		ClassID:   class.ID,
		Data:      fields.NewHstore(),
		Files:     fields.NewHstore(),
		Revision:  1,
		CreatedAt: fields.NewTime(&now),
		UpdatedAt: fields.NewTime(&now),
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
func (m *DataObject) BeforeUpdate(ctx context.Context) (context.Context, error) {
	m.UpdatedAt.Set(time.Now()) // nolint: errcheck
	return ctx, nil
}
