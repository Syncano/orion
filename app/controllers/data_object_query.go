package controllers

import (
	"fmt"
	"net/http"

	"github.com/go-pg/pg/v9/orm"
	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
	"github.com/Syncano/orion/app/query"
	"github.com/Syncano/orion/app/settings"
	"github.com/Syncano/orion/pkg/storage"
)

func newQueryError(detail string) *api.Error {
	return api.NewError(http.StatusBadRequest, map[string]interface{}{"query": detail})
}

type DataObjectQuery struct {
	fields map[string]models.FilterField
}

func NewDataObjectQuery(fields map[string]models.FilterField) *DataObjectQuery {
	return &DataObjectQuery{fields: fields}
}

func (doq *DataObjectQuery) Parse(qf *query.Factory, c echo.Context, q *orm.Query) (*orm.Query, error) {
	qs := c.QueryParam("query")
	if qs == "" {
		return q, nil
	}

	var (
		m map[string]interface{}
	)

	if err := json.Unmarshal([]byte(qs), &m); err != nil {
		return nil, newQueryError("Invalid JSON.")
	}

	if err := doq.Validate(m, true); err != nil {
		return nil, err
	}

	return doq.ParseMap(qf, c, q, m)
}

func (doq *DataObjectQuery) ParseMap(qf *query.Factory, c storage.DBContext, q *orm.Query, m map[string]interface{}) (*orm.Query, error) {
	var (
		f   models.FilterField
		ok  bool
		err error
	)

	for name, props := range m {
		f, ok = doq.fields[name]
		if !ok {
			return nil, newQueryError(fmt.Sprintf(`Invalid field name specified or missing filter index: "%s".`, name))
		}

		for lookup, data := range props.(map[string]interface{}) {
			if q, err = doq.fieldQuery(qf, c, q, f, lookup, data); err != nil {
				return nil, err
			}
		}
	}

	return q, nil
}

func (doq *DataObjectQuery) Validate(m map[string]interface{}, top bool) error {
	var (
		ok        bool
		propMap   map[string]interface{}
		nestCount int
	)

	for name, props := range m {
		propMap, ok = props.(map[string]interface{})
		if !ok {
			return newQueryError(fmt.Sprintf(`Expected dict at "%s".`, name))
		}

		if _, ok = propMap["_is"]; ok {
			nestCount++
		}
	}

	if top {
		if nestCount > settings.API.DataObjectNestedQueriesMax {
			return newQueryError(fmt.Sprintf("Too many nested queries defined (exceeds %d).", settings.API.DataObjectNestedQueriesMax))
		}
	} else {
		if nestCount > 0 {
			return newQueryError("Double nested queries are not allowed.")
		}
	}

	return nil
}

func (doq *DataObjectQuery) fieldQuery(qf *query.Factory, c storage.DBContext, q *orm.Query, f models.FilterField, lookup string, data interface{}) (*orm.Query, error) {
	// Find supported filter.
	if filts, ok := filters[lookup]; ok {
		for _, filt := range filts {
			if filt.Supports(f) {
				return filt.Process(qf, c, doq, q, f, lookup, data)
			}
		}
	}

	return nil, newQueryError(fmt.Sprintf(`Invalid lookup "%s" defined for field "%s".`, lookup, f.Name()))
}
