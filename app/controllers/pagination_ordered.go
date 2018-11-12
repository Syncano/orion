package controllers

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/labstack/echo"

	"github.com/Syncano/orion/app/api"
	"github.com/Syncano/orion/app/models"
)

const orderByQuery = "order_by"

type keysetcursorObject struct {
	id  int
	val interface{}
}

type keysetcursor struct {
	*cursor
	field     models.OrderField
	orderBy   string
	lastValue interface{}
}

func (c *keysetcursor) buildURL(path string, direction int, o interface{}) string {
	lastPk := c.lastPk
	lastVal := c.lastValue
	if o != nil {
		obj := o.(keysetcursorObject)
		lastPk = obj.id
		lastVal = obj.val
	}

	base := fmt.Sprintf("%s?direction=%d&order_by=%s", path, direction, c.orderBy)
	if c.limit != defaultLimit {
		base += fmt.Sprintf("&page_size=%d", c.limit)
	}
	if lastPk != 0 {
		base += fmt.Sprintf("&last_pk=%d", lastPk)
	}
	if v, err := c.field.ToString(lastVal); err == nil {
		base += fmt.Sprintf("&last_value=%s", url.QueryEscape(v))
	}
	return base
}

// NextURL ...
func (c *keysetcursor) NextURL(path string) string {
	o := c.last
	if !c.forward {
		o = c.first
	}
	return c.buildURL(path, 1, o)
}

// PrevURL ...
func (c *keysetcursor) PrevURL(path string) string {
	o := c.first
	if !c.forward {
		o = c.last
	}
	return c.buildURL(path, 0, o)
}

// PaginatorOrderedDB ...
type PaginatorOrderedDB struct {
	*PaginatorDB
	OrderFields map[string]models.OrderField
}

// FilterObjects ...
func (p *PaginatorOrderedDB) FilterObjects(cursor Cursorer) error {
	q := p.Query
	isOrderAsc := cursor.IsOrderAsc()
	limit := cursor.Limit()

	cur := cursor.(*keysetcursor)

	if cur.field == nil {
		return api.NewBadRequestError(`Missing or unindexed field used as "order_by".`)
	}

	sqlName := cur.field.SQLName()

	if cur.lastPk > 0 {
		if cur.forward {
			if cur.lastValue == nil {
				// Nulls Last, so if we got null value, find more nulls that have higher pk.
				// field IS NULL AND pk > last_pk
				q = q.Where(fmt.Sprintf("%s IS NULL AND ?TableAlias.id > ?", sqlName),
					cur.lastPk)
			} else {
				// Null value not yet reached, look for null values or with higher pair.
				// field IS NULL OR (field, pk) > (last_val, last_pk)
				q = q.Where(fmt.Sprintf("%s IS NULL OR (%s, ?TableAlias.id) > (?, ?)", sqlName, sqlName),
					cur.lastValue, cur.lastPk)
			}
		} else {
			if cur.lastValue == nil {
				// Null value reached, look for values that are not null or other nulls with lower pk.
				// field IS NOT NULL OR (field IS NULL AND pk < last_pk)
				q = q.Where(fmt.Sprintf("%s IS NOT NULL OR (%s IS NULL AND ?TableAlias.id < ?)", sqlName, sqlName),
					cur.lastPk)
			} else {
				// Last value was not null, so proceed normally.
				// (field, pk) < (last_val, last_pk)
				q = q.Where(fmt.Sprintf("(%s, ?TableAlias.id) < (?, ?)", sqlName),
					cur.lastValue, cur.lastPk)
			}
		}
	}

	order := orderingMap[isOrderAsc]
	p.Query = q.OrderExpr(fmt.Sprintf("%s %s, ?TableAlias.id %s", sqlName, order, order)).Limit(limit)
	return nil
}

// CreateCursor ...
func (p *PaginatorOrderedDB) CreateCursor(c echo.Context, defaultOrderAsc bool) Cursorer {
	cur := &keysetcursor{cursor: newCursor(c, defaultOrderAsc)}

	// Ignore default ordering.
	orderAsc := true
	orderBy := c.QueryParam(orderByQuery)
	cur.orderBy = orderBy

	if len(orderBy) > 0 && orderBy[0] == '-' {
		orderAsc = false
		orderBy = orderBy[1:]
	}

	for name, f := range p.OrderFields {
		if orderBy == name {
			cur.field = f
			break
		}
	}

	if s := c.QueryParam("last_value"); s != "" && cur.field != nil {
		cur.lastValue, _ = cur.field.FromString(s)
	}

	// Process order. Reverse order if direction is not forward.
	if !cur.forward {
		orderAsc = !orderAsc
	}
	cur.orderAsc = orderAsc

	cur.extractor = func(v interface{}) interface{} {
		return keysetcursorObject{
			id:  int(reflect.ValueOf(v).Elem().FieldByName("ID").Int()),
			val: cur.field.Get(v),
		}
	}
	return cur
}
