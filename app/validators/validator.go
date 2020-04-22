package validators

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg/v9/orm"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// sql_select/notexists/exists: optional param1: field (default "id"), optional param2: structquery (default: structfield+"Q")
// sql_select: make sure <structquery>.Where(<field>=this_value).Select() returns no error
// sql_exists: make sure <structquery>.Where(<field>=this_value).Exists() returns true
// sql_notexists: make sure <structquery>.Where(<field>=this_value).Exists() returns false

func init() {
	initTranslator()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	for _, v := range []struct {
		tag           string
		validatorFunc validator.Func
	}{
		{
			tag:           "sql_exists",
			validatorFunc: sqlExistsFunc,
		},
		{
			tag:           "sql_notexists",
			validatorFunc: sqlNotExistsFunc,
		},
		{
			tag:           "sql_select",
			validatorFunc: sqlSelectFunc,
		},
	} {
		validate.RegisterValidation(v.tag, v.validatorFunc) // nolint: errcheck
	}
}

func sqlFuncGetQuery(fl validator.FieldLevel) *orm.Query {
	field := "id"
	structquery := fl.StructFieldName() + "Q"
	params := strings.Split(fl.Param(), " ")

	if len(params) >= 1 && params[0] != "" {
		field = params[0]
	}

	if len(params) >= 2 && params[1] != "" {
		structquery = params[1]
	}

	q := fl.Parent().Elem().FieldByName(structquery).Interface().(*orm.Query)

	return q.Where(fmt.Sprintf("?TableAlias.%s = ?", field), fl.Field().Interface())
}

func sqlExistsFunc(fl validator.FieldLevel) bool {
	if exists, err := sqlFuncGetQuery(fl).Exists(); err == nil {
		return exists
	}

	return false
}

func sqlSelectFunc(fl validator.FieldLevel) bool {
	return sqlFuncGetQuery(fl).Select() == nil
}

func sqlNotExistsFunc(fl validator.FieldLevel) bool {
	if exists, err := sqlFuncGetQuery(fl).Exists(); err == nil {
		return !exists
	}

	return false
}

// Validator is a struct for echo.Validator.
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates new validator satisfying echo.Validator interface.
func NewValidator() *Validator {
	return &Validator{validator: validate}
}

// Validate validates object against validation library.
func (cv *Validator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
