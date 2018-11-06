package models

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-pg/pg/orm"
	"github.com/jackc/pgx/pgtype"
	json "github.com/json-iterator/go"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkbhex"

	"github.com/Syncano/orion/pkg/util"
)

var defaultObjectOrderFields = map[string]OrderField{
	"created_at": NewObjectField((*DataObject)(nil), "", "created_at", FieldDatetimeType),
	"updated_at": NewObjectField((*DataObject)(nil), "", "updated_at", FieldDatetimeType),
}

var defaultObjectFilterFields = map[string]FilterField{
	"id":         NewObjectField((*DataObject)(nil), "", "id", FieldIntegerType),
	"created_at": NewObjectField((*DataObject)(nil), "", "created_at", FieldDatetimeType),
	"updated_at": NewObjectField((*DataObject)(nil), "", "updated_at", FieldDatetimeType),
	"revision":   NewObjectField((*DataObject)(nil), "", "revision", FieldIntegerType),
}

var defaultUserOrderFields = map[string]OrderField{
	"created_at": NewObjectField((*DataObject)(nil), "profile", "created_at", FieldDatetimeType),
	"updated_at": NewObjectField((*DataObject)(nil), "profile", "updated_at", FieldDatetimeType),
}

var defaultUserFilterFields = map[string]FilterField{
	"id":         NewObjectField((*User)(nil), "", "id", FieldIntegerType),
	"username":   NewObjectField((*User)(nil), "", "username", FieldStringType),
	"created_at": NewObjectField((*DataObject)(nil), "profile", "created_at", FieldDatetimeType),
	"updated_at": NewObjectField((*DataObject)(nil), "profile", "updated_at", FieldDatetimeType),
	"revision":   NewObjectField((*DataObject)(nil), "profile", "revision", FieldIntegerType),
}

// Field types for object fields.
const (
	FieldStringType    = "string"
	FieldTextType      = "text"
	FieldIntegerType   = "integer"
	FieldFloatType     = "float"
	FieldBooleanType   = "boolean"
	FieldDatetimeType  = "datetime"
	FieldFileType      = "file"
	FieldReferenceType = "reference"
	FieldRelationType  = "relation"
	FieldObjectType    = "object"
	FieldArrayType     = "array"
	FieldGeopointType  = "geopoint"

	pgTimestamptzMinuteFormat = "2006-01-02 15:04:05.999999999Z07:00"
)

// ErrNilValue is used to signal that value passed was nil.
var ErrNilValue = errors.New("nil value")

// ValueFromString returns field's internal type object from string param.
// nolint: gocyclo
func ValueFromString(typ, s string) (interface{}, error) {
	switch typ {
	case FieldStringType:
		fallthrough
	case FieldTextType:
		return s, nil

	case FieldIntegerType:
		return strconv.Atoi(s)

	case FieldFloatType:
		return strconv.ParseFloat(s, 64)

	case FieldBooleanType:
		return util.IsTrue(s), nil

	case FieldDatetimeType:
		return time.Parse(pgTimestamptzMinuteFormat, s)

	case FieldFileType:
		// TODO: CORE-2468 files support
		return s, nil

	case FieldReferenceType:
		return strconv.Atoi(s)

	case FieldRelationType:
		var i []int
		err := json.Unmarshal([]byte("["+s[1:len(s)-1]+"]"), &i)
		return i, err

	case FieldObjectType:
		fallthrough
	case FieldArrayType:
		var i interface{}
		return i, json.Unmarshal([]byte(s), &i)

	case FieldGeopointType:
		return ewkbhex.Decode(s)
	}
	return nil, nil
}

// ValueToString converts value to pg string (to e.g. store it in hstore).
// nolint: gocyclo
func ValueToString(typ string, val interface{}) (string, error) {
	if val == nil {
		return "", ErrNilValue
	}

	switch typ {
	case FieldStringType:
		fallthrough
	case FieldTextType:
		return val.(string), nil

	case FieldIntegerType:
		return strconv.Itoa(val.(int)), nil

	case FieldFloatType:
		return strconv.FormatFloat(val.(float64), 'f', -1, 64), nil

	case FieldBooleanType:
		if val.(bool) {
			return "true", nil
		}
		return "false", nil

	case FieldDatetimeType:
		return val.(Time).Time.UTC().Format(pgTimestamptzMinuteFormat), nil

	case FieldFileType:
		// TODO: CORE-2468 file processing
		return val.(string), nil

	case FieldReferenceType:
		return strconv.Itoa(val.(int)), nil

	case FieldRelationType:
		vArr := val.([]int)
		ids := make([]string, len(vArr))
		for i, k := range vArr {
			ids[i] = strconv.Itoa(k)
		}

		sb := strings.Builder{}
		sb.WriteByte('{')
		sb.WriteString(strings.Join(ids, ","))
		sb.WriteByte('}')
		return sb.String(), nil

	case FieldObjectType:
		fallthrough
	case FieldArrayType:
		if v, err := json.Marshal(val); err == nil {
			return string(v), nil
		}

	case FieldGeopointType:
		return ewkbhex.Encode(val.(*geom.Point), binary.LittleEndian)
	}

	return "", nil
}

// SimpleObjectField ...
type SimpleObjectField struct {
	name  string
	table string
	typ   string
	field *orm.Field
}

// Get returns field's internal type object from object.
func (f *SimpleObjectField) Get(o interface{}) interface{} {
	return f.field.Value(reflect.ValueOf(o).Elem()).Interface()
}

// ToString ...
func (f *SimpleObjectField) ToString(v interface{}) (string, error) {
	return ValueToString(f.typ, v)
}

// FromString ...
func (f *SimpleObjectField) FromString(s string) (interface{}, error) {
	return ValueFromString(f.typ, s)
}

// Name ...
func (f *SimpleObjectField) Name() string {
	return f.name
}

// Type ...
func (f *SimpleObjectField) Type() string {
	return f.typ
}

// SQLName ...
func (f *SimpleObjectField) SQLName() string {
	return fmt.Sprintf("%s.%s", f.table, f.field.SQLName)
}

// NewObjectField ...
func NewObjectField(m interface{}, alias, fieldName, typ string) *SimpleObjectField {
	table := orm.GetTable(reflect.TypeOf(m).Elem())
	if len(alias) == 0 {
		alias = string(table.Alias)
	}
	return &SimpleObjectField{name: fieldName, typ: typ, field: table.FieldsMap[fieldName], table: alias}
}

// DataObjectField is used to define each field in data object schema.
type DataObjectField struct {
	FName       string `mapstructure:"name"`
	FType       string `mapstructure:"type"`
	OrderIndex  bool   `mapstructure:"order_index"`
	FilterIndex bool   `mapstructure:"filter_index"`
	Unique      bool   `mapstructure:"unique"`
	Target      string `mapstructure:"target"`

	TableAlias string
	Mapping    string
}

// Name ...
func (f *DataObjectField) Name() string {
	return f.FName
}

// Type ...
func (f *DataObjectField) Type() string {
	return f.FType
}

// ToString ...
func (f *DataObjectField) ToString(v interface{}) (string, error) {
	return ValueToString(f.FType, v)
}

// FromString ...
func (f *DataObjectField) FromString(s string) (interface{}, error) {
	return ValueFromString(f.FType, s)
}

// SQLName ...
func (f *DataObjectField) SQLName() string {
	var typ string
	field := fmt.Sprintf(`("%s"."_data"->'%s')`, f.TableAlias, f.Mapping)

	switch f.FType {
	case FieldStringType:
		typ = "varchar(128)"
	case FieldTextType:
		typ = "text"
	case FieldIntegerType:
		typ = "integer"
	case FieldFloatType:
		typ = "double precision"
	case FieldBooleanType:
		typ = "boolean"
	case FieldDatetimeType:
		return "?schema.to_timestamp" + field
	case FieldFileType:
		typ = "varchar(100)"
	case FieldReferenceType:
		typ = "integer"
	case FieldRelationType:
		typ = "integer[]"
	case FieldObjectType:
		fallthrough
	case FieldArrayType:
		typ = "jsonb"
	case FieldGeopointType:
		typ = "geography(POINT,4326)"
	default:
		return f.FName
	}
	return field + "::" + typ
}

// Get returns field's internal type object from hstore.
func (f *DataObjectField) Get(o interface{}) interface{} {
	data := o.(*DataObject).Data
	val := data.Map[f.Mapping]
	if val.Status != pgtype.Present {
		return nil
	}
	if v, err := f.FromString(val.String); err == nil {
		return v
	}
	return nil
}

// Set formats field's internal type object to string and sets it to hstore.
func (f *DataObjectField) Set(data *Hstore, val interface{}) error {
	if val == nil {
		data.Map[f.Mapping] = pgtype.Text{Status: pgtype.Null}
		return nil
	}

	v, err := f.ToString(val)

	if err != nil {
		data.Map[f.Mapping] = pgtype.Text{Status: pgtype.Null}
		return err
	}
	data.Map[f.Mapping] = pgtype.Text{String: v, Status: pgtype.Present}
	return err
}
