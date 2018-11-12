package controllers

import (
	"reflect"

	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/pkg/redisdb"
)

// PaginatorRedis ...
type PaginatorRedis struct {
	DBCtx         *redisdb.DBCtx
	SkippedFields []string
}

// FilterObjects ...
func (p *PaginatorRedis) FilterObjects(cursor Cursorer) error {
	lastPk := cursor.LastPK()
	isOrderAsc := cursor.IsOrderAsc()
	limit := cursor.Limit()
	var minPK, maxPK int

	if lastPk != 0 {
		if isOrderAsc {
			minPK = lastPk
		} else {
			maxPK = lastPk
		}
	}

	return p.DBCtx.List(minPK, maxPK, limit, isOrderAsc, p.SkippedFields)
}

// ProcessObjects ...
func (p *PaginatorRedis) ProcessObjects(c echo.Context, cursor Cursorer, typ reflect.Type, serializer serializers.Serializer, responseLimit *int) ([]api.RawMessage, error) {
	var ret []api.RawMessage

	r := p.DBCtx.Value()

	var (
		data      []byte
		e         error
		obj, last interface{}
	)
	for i := 0; i < r.Len(); i++ {
		obj = r.Index(i).Interface()
		data, e = api.Marshal(c, serializer.Response(obj))
		if last == nil {
			cursor.SetFirst(obj)
		}
		last = obj
		ret = append(ret, data)
		if e != nil {
			return nil, e
		}
	}
	if last != nil {
		cursor.SetLast(obj)
	}

	return ret, nil
}

// CreateCursor ...
func (p *PaginatorRedis) CreateCursor(c echo.Context, defaultOrderAsc bool) Cursorer {
	return newCursor(c, defaultOrderAsc)
}
