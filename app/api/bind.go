package api

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

// Binder ...
type Binder struct{}

// Bind implements the `Binder#Bind` function.
func (b *Binder) Bind(i interface{}, c echo.Context) error {
	req := c.Request()
	if req.ContentLength == 0 {
		if req.Method == echo.GET || req.Method == echo.DELETE {
			return nil
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Request body can't be empty")
	}
	ctype := req.Header.Get(echo.HeaderContentType)
	switch {
	case strings.HasPrefix(ctype, echo.MIMEApplicationForm), strings.HasPrefix(ctype, echo.MIMEMultipartForm):
		params, err := c.FormParams()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Error parsing form")
		}
		if err = b.bindData(c, i, params, "form"); err != nil {
			if _, ok := err.(*Error); ok {
				return err
			}
			e := echo.NewHTTPError(http.StatusBadRequest, "Error parsing form")
			e.Internal = err
			return e
		}
	default:
		return echo.ErrUnsupportedMediaType
	}
	return nil
}

func (b *Binder) bindData(c echo.Context, ptr interface{}, data map[string][]string, tag string) error { // nolint: gocyclo
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()

	if typ.Kind() != reflect.Struct {
		return errors.New("binding element must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}
		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get(tag)

		if inputFieldName == "" {
			inputFieldName = typeField.Name
			// If tag is nil, we inspect if the field is a struct.
			if _, ok := bindUnmarshaler(structField); !ok && structFieldKind == reflect.Struct {
				err := b.bindData(c, structField.Addr().Interface(), data, tag)
				if err != nil {
					return err
				}
				continue
			}
		}
		inputValue, exists := data[inputFieldName]
		if !exists {
			continue
		}

		// Call this first, in case we're dealing with an alias to an array type
		if ok, err := unmarshalField(typeField.Type.Kind(), inputValue[0], structField); ok {
			if err != nil {
				return err
			}
			continue
		}

		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for j := 0; j < numElems; j++ {
				if err := setWithProperType(c, sliceOf, inputValue[j], slice.Index(j)); err != nil {
					return NewError(http.StatusBadRequest, map[string]interface{}{
						inputFieldName: fmt.Sprintf("Incorrect type. Expected %s.", sliceOf)})
				}
			}
			val.Field(i).Set(slice)
		} else {
			typ := typeField.Type.Kind()
			if err := setWithProperType(c, typ, inputValue[0], structField); err != nil {
				return NewError(http.StatusBadRequest, map[string]interface{}{
					inputFieldName: fmt.Sprintf("Incorrect type. Expected %s.", typ)})
			}
		}
	}
	return nil
}

func setWithProperType(c echo.Context, valueKind reflect.Kind, val string, structField reflect.Value) error { // nolint: gocyclo
	// But also call it here, in case we're dealing with an array of BindUnmarshalers
	if ok, err := unmarshalField(valueKind, val, structField); ok {
		return err
	}

	switch valueKind {
	case reflect.Ptr:
		if _, ok := structField.Interface().(*multipart.FileHeader); ok {
			return setFileheaderField(c, val, structField)
		}
		return setWithProperType(c, structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("unknown type")
	}
	return nil
}

func unmarshalField(valueKind reflect.Kind, val string, field reflect.Value) (bool, error) {
	switch valueKind {
	case reflect.Ptr:
		return unmarshalFieldPtr(val, field)
	default:
		return unmarshalFieldNonPtr(val, field)
	}
}

// bindUnmarshaler attempts to unmarshal a reflect.Value into a BindUnmarshaler
func bindUnmarshaler(field reflect.Value) (echo.BindUnmarshaler, bool) {
	ptr := reflect.New(field.Type())
	if ptr.CanInterface() {
		iface := ptr.Interface()
		if unmarshaler, ok := iface.(echo.BindUnmarshaler); ok {
			return unmarshaler, ok
		}
	}
	return nil, false
}

func unmarshalFieldNonPtr(value string, field reflect.Value) (bool, error) {
	if unmarshaler, ok := bindUnmarshaler(field); ok {
		err := unmarshaler.UnmarshalParam(value)
		field.Set(reflect.ValueOf(unmarshaler).Elem())
		return true, err
	}
	return false, nil
}

func unmarshalFieldPtr(value string, field reflect.Value) (bool, error) {
	if field.IsNil() {
		// Initialize the pointer to a nil value
		field.Set(reflect.New(field.Type().Elem()))
	}
	return unmarshalFieldNonPtr(value, field.Elem())
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func setFileheaderField(c echo.Context, value string, field reflect.Value) error {
	fh, err := c.FormFile(value)
	if err == nil {
		field.Set(reflect.ValueOf(fh))
	}
	return err
}
