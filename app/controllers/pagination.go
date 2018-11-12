package controllers

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-pg/pg/orm"
	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/serializers"
	"github.com/Syncano/orion/pkg/settings"
)

const (
	defaultLimit            = 100
	maxResponseLimit        = 2 * 1024 * 1024
	contextResponseLimitKey = "responseLimit"
)

var (
	orderingMap = map[bool]string{
		true:  "asc",
		false: "desc",
	}
	errorType        = reflect.TypeOf((*error)(nil)).Elem()
	errStopIteration = errors.New("stop iteration")
)

type cursorObject struct {
	id int
}

type cursor struct {
	limit           int
	forward         bool
	lastPk          int
	defaultOrderAsc bool
	orderAsc        bool

	extractor func(interface{}) interface{}
	first     interface{}
	last      interface{}
}

func newCursor(c echo.Context, defaultOrderAsc bool) *cursor {
	limit, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil || limit > settings.API.MaxPageSize || limit < 0 {
		limit = defaultLimit
	}

	forward := true
	if direction, err := strconv.Atoi(c.QueryParam("direction")); err == nil {
		forward = direction != 0
	}

	lastPk, _ := strconv.Atoi(c.QueryParam("last_pk"))

	// Process order. Reverse order if direction is not forward.
	orderAsc := defaultOrderAsc
	if ordering := c.QueryParam("ordering"); ordering != "" {
		orderAsc = strings.ToLower(ordering) == "asc"
	}
	if !forward {
		orderAsc = !orderAsc
		defaultOrderAsc = !defaultOrderAsc
	}

	extractor := func(v interface{}) interface{} {
		return cursorObject{id: int(reflect.ValueOf(v).Elem().FieldByName("ID").Int())}
	}

	return &cursor{limit: limit, forward: forward, lastPk: lastPk,
		defaultOrderAsc: defaultOrderAsc, orderAsc: orderAsc, extractor: extractor}
}

func (c *cursor) buildURL(path string, direction int, o interface{}) string {
	lastPk := c.lastPk
	if o != nil {
		lastPk = o.(cursorObject).id
	}

	base := fmt.Sprintf("%s?direction=%d", path, direction)
	if c.limit != defaultLimit {
		base += fmt.Sprintf("&page_size=%d", c.limit)
	}
	if lastPk != 0 {
		base += fmt.Sprintf("&last_pk=%d", lastPk)
	}
	if c.orderAsc != c.defaultOrderAsc {
		base += fmt.Sprintf("&ordering=%s", orderingMap[c.orderAsc])
	}
	return base
}

func (c *cursor) Limit() int {
	return c.limit
}
func (c *cursor) LastPK() int {
	return c.lastPk
}
func (c *cursor) IsOrderAsc() bool {
	return c.orderAsc
}
func (c *cursor) IsForward() bool {
	return c.forward
}
func (c *cursor) SetFirst(v interface{}) {
	c.first = c.extractor(v)
}
func (c *cursor) SetLast(v interface{}) {
	c.last = c.extractor(v)
}

// NextURL ...
func (c *cursor) NextURL(path string) string {
	o := c.last
	if !c.forward {
		o = c.first
	}
	return c.buildURL(path, 1, o)
}

// PrevURL ...
func (c *cursor) PrevURL(path string) string {
	o := c.first
	if !c.forward {
		o = c.last
	}
	return c.buildURL(path, 0, o)
}

// PaginatorDB ...
type PaginatorDB struct {
	Query *orm.Query
}

// FilterObjects ...
func (p *PaginatorDB) FilterObjects(cursor Cursorer) error {
	// Filter by ID.
	q := p.Query
	lastPk := cursor.LastPK()
	isOrderAsc := cursor.IsOrderAsc()
	limit := cursor.Limit()

	if lastPk != 0 {
		if isOrderAsc {
			q = q.Where("?TableAlias.id > ?", lastPk)
		} else {
			q = q.Where("?TableAlias.id < ?", lastPk)
		}
	}

	p.Query = q.OrderExpr("?TableAlias.id " + orderingMap[isOrderAsc]).Limit(limit)
	return nil
}

// ProcessObjects ...
func (p *PaginatorDB) ProcessObjects(c echo.Context, cursor Cursorer, typ reflect.Type, serializer serializers.Serializer, responseLimit *int) ([]api.RawMessage, error) {
	var ret []api.RawMessage
	q := p.Query

	// Create foreach function using reflection.
	var (
		resp      []byte
		e         error
		obj, last interface{}
	)
	foreach := reflect.MakeFunc(reflect.FuncOf([]reflect.Type{typ}, []reflect.Type{errorType}, false),
		func(args []reflect.Value) (results []reflect.Value) {
			if *responseLimit <= 0 {
				return []reflect.Value{reflect.ValueOf(&errStopIteration).Elem()}
			}

			obj = args[0].Interface()
			resp, e = api.Marshal(c, serializer.Response(obj))
			if e != nil {
				return []reflect.Value{reflect.ValueOf(&e).Elem()}
			}

			if last == nil {
				cursor.SetFirst(obj)
			}
			last = obj

			ret = append(ret, resp)
			*responseLimit -= len(resp)

			return []reflect.Value{reflect.Zero(errorType)}
		})

	if err := q.ForEach(foreach.Interface()); err != nil && err != errStopIteration {
		return nil, err
	}
	if last != nil {
		cursor.SetLast(last)
	}

	return ret, nil
}

// CreateCursor ...
func (p *PaginatorDB) CreateCursor(c echo.Context, defaultOrderAsc bool) Cursorer {
	return newCursor(c, defaultOrderAsc)
}

// Paginate ...
// nolint: gocyclo
func Paginate(c echo.Context, cursor Cursorer, typ interface{}, serializer serializers.Serializer, paginator Paginator) ([]api.RawMessage, error) {
	if cursor.Limit() == 0 {
		return []api.RawMessage{}, nil
	}

	// Use pointer to int so that it is modified in loop instead of being copied.
	var responseLimit *int

	if v := c.Get(contextResponseLimitKey); v != nil {
		responseLimit = v.(*int)
	} else {
		m := maxResponseLimit
		responseLimit = &m
	}

	if err := paginator.FilterObjects(cursor); err != nil {
		return nil, err
	}
	ret, err := paginator.ProcessObjects(c, cursor, reflect.TypeOf(typ), serializer, responseLimit)
	if err != nil {
		return nil, err
	}

	// Reverse required if direction is not forward.
	rLen := len(ret)
	if !cursor.IsForward() {
		for i := 0; i < rLen/2; i++ {
			ret[i], ret[rLen-1-i] = ret[rLen-1-i], ret[i]
		}
	}

	// Process has next/prev.
	var hasNext, hasPrev bool

	limitReached := rLen == cursor.Limit() && cursor.Limit() > 0
	if cursor.IsForward() {
		hasNext = limitReached
		hasPrev = cursor.LastPK() > 0
	} else {
		hasPrev = limitReached
		hasNext = cursor.LastPK() > 0
	}

	req := c.Request()
	if hasNext {
		c.Set("next", cursor.NextURL(req.URL.Path))
	}
	if hasPrev {
		c.Set("prev", cursor.PrevURL(req.URL.Path))
	}
	return ret, nil
}
