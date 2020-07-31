package controllers

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/now"
	json "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"

	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/pkg-go/v2/database"
)

var filters = map[string][]*filterOp{}

const (
	maxListLength = 128
)

func registerFilter(op *filterOp, lookups ...string) {
	for _, lookup := range lookups {
		filters[lookup] = append(filters[lookup], op)
	}
}

// FilterOp represents every filter functionality and validation.
type filterOp struct {
	supportedTypes    []string
	unsupportedTypes  []string
	expectedValue     []reflect.Kind
	expectList        bool
	expectedListValue []reflect.Kind
	query             func(qf *query.Factory, c database.DBContext, q *orm.Query, f models.FilterField, op string, data interface{}) *orm.Query
	validate          func(qf *query.Factory, c database.DBContext, q *filterOp, f models.FilterField, val interface{}) (interface{}, error)
}

func (op *filterOp) Supports(f models.FilterField) bool {
	typ := f.Type()

	if op.supportedTypes != nil {
		for _, t := range op.supportedTypes {
			if t == typ {
				return true
			}
		}

		return false
	}

	if op.unsupportedTypes != nil {
		for _, t := range op.unsupportedTypes {
			if t == typ {
				return false
			}
		}
	}

	return true
}

func (op *filterOp) Process(qf *query.Factory, c database.DBContext, doq *DataObjectQuery, q *orm.Query, f models.FilterField, lookup string, data interface{}) (*orm.Query, error) {
	var ok bool

	if op.expectList {
		// Validate list.
		ok = op.validateList(f, op.expectedListValue, data)
	} else {
		// Validate single value.
		data, ok = op.validateValue(f, op.expectedValue, data)
	}

	if !ok {
		return nil, newQueryError(fmt.Sprintf(`Invalid value type provided for "%s" lookup of field "%s".`, lookup, f.Name()))
	}

	if ok && op.validate != nil {
		var err error

		data, err = op.validate(qf, c, op, f, data)
		if err != nil {
			return nil, err
		}
	}

	if !ok || data == nil {
		return nil, newQueryError(fmt.Sprintf(`Validation of value provided "%s" lookup of field "%s" failed.`, lookup, f.Name()))
	}

	return op.query(qf, c, q, f, lookup, data), nil
}

func (op *filterOp) validateKind(k reflect.Kind, expected []reflect.Kind, val interface{}) (interface{}, bool) {
	for _, v := range expected {
		if k == reflect.Float64 && v == reflect.Int {
			return int(val.(float64)), true
		}

		if k == v {
			return val, true
		}
	}

	return nil, false
}

func (op *filterOp) validateList(f models.FilterField, expectedListValue []reflect.Kind, lst interface{}) bool {
	if lst == nil {
		return false
	}

	kind := reflect.TypeOf(lst).Kind()
	if kind != reflect.Slice {
		return false
	}

	// Iterate through slice and validate each value (converting it if needed).
	arr := lst.([]interface{})
	if len(arr) == 0 || len(arr) > maxListLength {
		return false
	}

	var (
		val interface{}
		ok  bool
	)

	for i, v := range arr {
		val, ok = op.validateValue(f, expectedListValue, v)
		if !ok {
			return false
		}

		arr[i] = val
	}

	return true
}

func (op *filterOp) validateValue(f models.FilterField, expectedValue []reflect.Kind, val interface{}) (interface{}, bool) {
	if val == nil {
		return nil, false
	}

	kind := reflect.TypeOf(val).Kind()

	if len(expectedValue) > 0 {
		return op.validateKind(kind, expectedValue, val)
	}

	// Use default mapping.
	switch f.Type() {
	case models.FieldStringType:
		return val, kind == reflect.String

	case models.FieldReferenceType, models.FieldIntegerType:
		if kind == reflect.Float64 {
			return int(val.(float64)), true
		}

	case models.FieldFloatType:
		return val, kind == reflect.Float64

	case models.FieldBooleanType:
		return val, kind == reflect.Bool

	case models.FieldDatetimeType:
		if kind == reflect.String {
			if t, err := now.ParseInLocation(time.UTC, val.(string)); err == nil {
				return t, true
			}
		}
	}

	return nil, false
}

// nolint: gocyclo
func init() {
	validate := validator.New()

	// Common filter - exists.
	registerFilter(&filterOp{
		expectedValue:    []reflect.Kind{reflect.Bool},
		unsupportedTypes: []string{models.FieldArrayType},
		query: func(qf *query.Factory, c database.DBContext, q *orm.Query, f models.FilterField, op string, data interface{}) *orm.Query {
			sqlOp := "IS NULL"
			if data.(bool) {
				sqlOp = "IS NOT NULL"
			}
			return q.Where(fmt.Sprintf("%s %s", f.SQLName(), sqlOp))
		}},
		"_exists",
	)

	// Common filters - gt, gte, lt, lte, eq, neq.
	var simpleLookups = map[string]string{
		"_gt":  ">",
		"_gte": ">=",
		"_lt":  "<",
		"_lte": "<=",
		"_eq":  "=",
		"_neq": "!=",
	}

	registerFilter(&filterOp{
		unsupportedTypes: []string{models.FieldRelationType, models.FieldArrayType, models.FieldGeopointType},
		query: func(qf *query.Factory, c database.DBContext, q *orm.Query, f models.FilterField, op string, data interface{}) *orm.Query {
			return q.Where(fmt.Sprintf("%s %s ?", f.SQLName(), simpleLookups[op]), data)
		}},
		"_gt", "_gte", "_lt", "_lte", "_eq", "_neq",
	)

	// String specific filters (LIKE/ILIKE).
	var stringLookupFormats = map[string]string{
		"contains":   "%%%s%%",
		"startswith": "%s%%",
		"endswith":   "%%%s",
		"like":       "%s'",
		"eq":         "%s",
	}

	registerFilter(&filterOp{
		supportedTypes: []string{models.FieldStringType},
		query: func(qf *query.Factory, c database.DBContext, q *orm.Query, f models.FilterField, op string, data interface{}) *orm.Query {
			idx := 1
			sqlOp := "LIKE"
			if strings.HasPrefix(op, "_i") {
				idx = 2
				sqlOp = "ILIKE"
			}
			return q.Where(fmt.Sprintf("%s %s ?", f.SQLName(), sqlOp), fmt.Sprintf(stringLookupFormats[op[idx:]], data))
		}},
		"_contains", "_icontains", "_startswith", "_istartswith", "_endswith", "_iendswith", "_like", "_ilike", "_ieq",
	)

	// Container filters - in, nin.
	registerFilter(&filterOp{
		expectList:       true,
		unsupportedTypes: []string{models.FieldRelationType, models.FieldArrayType, models.FieldGeopointType},
		query: func(qf *query.Factory, c database.DBContext, q *orm.Query, f models.FilterField, op string, data interface{}) *orm.Query {
			sqlOp := "IN"
			if op == "_nin" {
				sqlOp = "NOT IN"
			}
			return q.Where(fmt.Sprintf("%s %s (?)", f.SQLName(), sqlOp), pg.In(data))
		}},
		"_in", "_nin",
	)

	// Relation contains.
	registerFilter(&filterOp{
		expectList:        true,
		expectedListValue: []reflect.Kind{reflect.Int},
		supportedTypes:    []string{models.FieldRelationType},
		query: func(qf *query.Factory, c database.DBContext, q *orm.Query, f models.FilterField, op string, data interface{}) *orm.Query {
			return q.Where(fmt.Sprintf("%s @> ?", f.SQLName()), pg.Array(data))
		}},
		"_contains",
	)

	// Array contains.
	registerFilter(&filterOp{
		expectList:        true,
		expectedListValue: []reflect.Kind{reflect.String, reflect.Bool, reflect.Float64, reflect.Int},
		supportedTypes:    []string{models.FieldArrayType},
		query: func(qf *query.Factory, c database.DBContext, q *orm.Query, f models.FilterField, op string, data interface{}) *orm.Query {
			arr, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			return q.Where(fmt.Sprintf("%s @> ?", f.SQLName()), string(arr))
		}},
		"_contains",
	)

	// Geo near.
	type nearLookup struct {
		Longitude            float64 `validate:"gt=-180,lt=180,required"`
		Latitude             float64 `validate:"gt=-90,lt=90,required"`
		DistanceInKilometers float64 `mapstructure:"distance_in_kilometers" validate:"gte=0"`
		DistanceInMiles      float64 `mapstructure:"distance_in_miles" validate:"gte=0"`
	}

	registerFilter(&filterOp{
		expectedValue:  []reflect.Kind{reflect.Map},
		supportedTypes: []string{models.FieldGeopointType},
		validate: func(qf *query.Factory, c database.DBContext, op *filterOp, f models.FilterField, val interface{}) (interface{}, error) {
			l := &nearLookup{}
			if mapstructure.Decode(val, l) != nil || validate.Struct(l) != nil {
				return nil, nil
			}

			if l.DistanceInMiles == 0 {
				l.DistanceInMiles = 100
			}
			if l.DistanceInKilometers == 0 {
				l.DistanceInKilometers = 1609.344 * l.DistanceInMiles
			}
			return l, nil
		},

		query: func(qf *query.Factory, c database.DBContext, q *orm.Query, f models.FilterField, op string, data interface{}) *orm.Query {
			l := data.(*nearLookup)
			return q.Where(fmt.Sprintf("ST_DWithin(%s, ST_GeomFromEWKB(?), ?)", f.SQLName()),
				&ewkb.Point{Point: geom.NewPointFlat(geom.XY, []float64{l.Longitude, l.Latitude})}, l.DistanceInKilometers)
		}},
		"_near",
	)

	// Reference and relation is.
	registerFilter(&filterOp{
		expectedValue:  []reflect.Kind{reflect.Map},
		supportedTypes: []string{models.FieldReferenceType, models.FieldRelationType},
		validate: func(qf *query.Factory, c database.DBContext, op *filterOp, f models.FilterField, val interface{}) (interface{}, error) {
			m, ok := val.(map[string]interface{})
			if !ok {
				return nil, nil
			}

			classMgr := qf.NewClassManager(c)
			dof := f.(*models.DataObjectField)
			cls := &models.Class{Name: dof.Target}

			// Users and Data Object are merged objects, use owner_id as id if targeting as user.
			col := "id"
			switch dof.Target {
			case "user":
				col = "owner_id"
			case "self":
				cls = c.Get(contextClassKey).(*models.Class)
			}
			if cls.ID == 0 && classMgr.OneByName(cls) != nil {
				return nil, newQueryError("Referenced class " + cls.Name + " does not exist.")
			}

			// Process subquery.
			var q *orm.Query
			switch cls.Name {
			case models.UserClassName:
				q = qf.NewUserManager(c).Query((*models.User)(nil)).
					Join(`JOIN ?schema.data_dataobject AS "profile" ON "profile"."owner_id" = "user"."id"`).
					Where("profile._klass_id = ?", cls.ID).Column(col)
			default:
				q = qf.NewDataObjectManager(c).ForClassQ(cls, (*models.DataObject)(nil)).Column(col)
			}

			doq := NewDataObjectQuery(cls.FilterFields())
			if err := doq.Validate(m, false); err != nil {
				return nil, err
			}
			q, err := doq.ParseMap(qf, c, q, m)
			if err != nil {
				return nil, err
			}

			return q.Limit(settings.API.DataObjectNestedQueryLimit), nil
		},

		query: func(qf *query.Factory, c database.DBContext, q *orm.Query, f models.FilterField, op string, data interface{}) *orm.Query {
			quer, err := data.(*orm.Query).AppendQuery(database.GetTenantDB(qf.Database(), c).Formatter(), nil)
			if err != nil {
				panic(err)
			}
			if f.Type() == models.FieldReferenceType {
				return q.Where(fmt.Sprintf("%s IN (%s)", f.SQLName(), string(quer)))
			}
			return q.Where(fmt.Sprintf("%s && ARRAY(%s)", f.SQLName(), string(quer)))
		}},
		"_is",
	)
}
