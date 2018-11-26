package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/Syncano/orion/pkg/settings"

	"github.com/jackc/pgx/pgtype"
	json "github.com/json-iterator/go"
)

var jsonNull = []byte("null")

// Time ...
type Time struct {
	pgtype.Timestamptz
}

// NewTime ...
func NewTime(t *time.Time) Time {
	if t == nil {
		now := time.Now()
		t = &now
	} else if t.IsZero() {
		return Time{Timestamptz: pgtype.Timestamptz{Status: pgtype.Null}}
	}
	return Time{Timestamptz: pgtype.Timestamptz{Time: t.UTC(), Status: pgtype.Present}}
}

// Value is used on value in go-pg, pass it to pointer version.
func (t Time) Value() (driver.Value, error) {
	return t.Timestamptz.Value()
}

func (t *Time) String() string {
	return t.Time.UTC().Format(settings.Common.DateTimeFormat)
}

// IsNull returns true if underlying value is null.
func (t *Time) IsNull() bool {
	return t.Status == pgtype.Null
}

// MarshalJSON ...
func (t *Time) MarshalJSON() ([]byte, error) {
	if t.IsNull() {
		return jsonNull, nil
	}
	return []byte(fmt.Sprintf("\"%s\"", t.String())), nil
}

// Date ...
type Date struct {
	pgtype.Date
}

// Value is used on value in go-pg, pass it to pointer version.
func (d Date) Value() (driver.Value, error) {
	return d.Date.Value()
}

// IsNull returns true if underlying value is null.
func (d *Date) IsNull() bool {
	return d.Status == pgtype.Null
}

func (d *Date) String() string {
	return d.Time.UTC().Format(settings.Common.DateTimeFormat)
}

// MarshalJSON ...
func (d *Date) MarshalJSON() ([]byte, error) {
	if d.IsNull() {
		return jsonNull, nil
	}
	return []byte(fmt.Sprintf("\"%s\"", d.String())), nil
}

// Daterange ...
type Daterange struct {
	pgtype.Daterange
}

// IsNull returns true if underlying value is null.
func (r *Daterange) IsNull() bool {
	return r.Status == pgtype.Null
}

// JSON ...
type JSON struct {
	pgtype.JSON
	Data interface{}
}

// Value implements the database/sql/driver Valuer interface.
func (j JSON) Value() (driver.Value, error) {
	if j.Data != nil {
		b, e := json.Marshal(j.Data)
		return string(b), e
	}
	return j.JSON.Value()
}

// Get ...
func (j *JSON) Get() interface{} {
	if j.Data == nil {
		j.Data = j.JSON.Get()
	}
	return j.Data
}

// Scan implements the database/sql Scanner interface.
func (j *JSON) Scan(src interface{}) error {
	err := j.JSON.Scan(src)
	j.Data = nil
	return err
}

// IsNull returns true if underlying value is null.
func (j *JSON) IsNull() bool {
	return j.JSON.Status == pgtype.Null
}

// MarshalJSON ...
func (j *JSON) MarshalJSON() ([]byte, error) {
	if j.Data != nil {
		return json.Marshal(j.Data)
	}
	return j.Bytes, nil
}

// Hstore ...
type Hstore struct {
	pgtype.Hstore
}

// NewHstore ...
func NewHstore() Hstore {
	return Hstore{Hstore: pgtype.Hstore{Map: make(map[string]pgtype.Text), Status: pgtype.Present}}
}

// Value is used on value in go-pg, pass it to pointer version.
func (h Hstore) Value() (driver.Value, error) {
	return h.Hstore.Value()
}

// IsNull returns true if underlying value is null.
func (h *Hstore) IsNull() bool {
	return h.Status == pgtype.Null
}
