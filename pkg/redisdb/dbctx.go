package redisdb

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/Syncano/orion/pkg/util"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

var (
	// ErrNotFound marks that Redis model was not found.
	ErrNotFound = errors.New("redis: object not found")
	// ErrExpectedMismatch marks that Update expected conditions wasn't met.
	ErrExpectedMismatch = errors.New("redis: expected mismatch")
)

const (
	updateRetries = 3
)

// DBCtx represents DB context.
type DBCtx struct {
	*DB
	args map[string]interface{}

	model Modeler
	table *Table
	value reflect.Value
}

// Find selects one object with specified pk.
func (c *DBCtx) Find(pk int) error {
	if c.value.Kind() != reflect.Struct {
		panic("redis: model is not a struct")
	}
	objectKey := c.getObjectKey(pk)
	r, err := c.redisCli.HGetAll(objectKey).Result()
	if err != nil {
		return err
	}
	if len(r) == 0 {
		return ErrNotFound
	}

	val := c.value
	var (
		name, v string
		ok      bool
		vv      reflect.Value
	)
	for _, f := range c.table.Fields {
		name = f.Name
		v, ok = r[name]
		if ok {
			vv = reflect.ValueOf(f.Adapter.Load(v))
		} else {
			if !f.HasDefault() {
				continue
			}
			vv = f.Default()
		}

		f.Value(val).Set(vv)
	}

	return nil
}

// Value ...
func (c *DBCtx) Value() reflect.Value {
	return c.value
}

func (c *DBCtx) listKeys(minPK, maxPK, limit int, isOrderAsc bool) ([]string, error) {
	var (
		min, max string
	)
	if maxPK == 0 {
		max = "+inf"
	} else {
		max = strconv.Itoa(maxPK)
	}
	if minPK == 0 {
		min = "-inf"
	} else {
		min = strconv.Itoa(minPK)
	}

	l := c.model.ListMaxSize()
	listKey := c.getListKey()
	if limit > l {
		limit = l
	}

	opt := redis.ZRangeBy{Max: max, Min: min, Count: int64(limit)}
	if isOrderAsc {
		return c.redisCli.ZRangeByScore(listKey, opt).Result()
	}
	return c.redisCli.ZRevRangeByScore(listKey, opt).Result()
}

func (c *DBCtx) createSlice(objs []reflect.Value) {
	sv := reflect.MakeSlice(c.value.Type(), len(objs), len(objs))
	for i, o := range objs {
		sv.Index(i).Set(o)
	}
	c.value.Set(sv)
	c.value = sv
}

func (c *DBCtx) getFields(included, skipped []string) []string {
	skip := make(map[string]struct{})
	for _, f := range skipped {
		skip[f] = struct{}{}
	}

	incl := make(map[string]struct{})
	if len(included) == 0 {
		for n := range c.table.Fields {
			incl[n] = struct{}{}
		}
	} else {
		for _, f := range included {
			incl[f] = struct{}{}
		}
	}

	var (
		fields []string
	)
	for n := range c.table.Fields {
		if _, ok := skip[n]; ok {
			continue
		}
		if _, ok := incl[n]; ok {
			fields = append(fields, n)
		}
	}
	return fields
}

func (c *DBCtx) getObjectKey(pk int) string {
	return fmt.Sprintf("%s:%d", c.model.Key(c.args), pk)
}

func (c *DBCtx) getListKey() string {
	return fmt.Sprintf("%s:set:%s", c.model.Key(c.args), c.model.ListArgs(c.args))
}

// List selects a list of objects.
func (c *DBCtx) List(minPK, maxPK, limit int, isOrderAsc bool, skippedFields []string) error {
	if c.value.Kind() != reflect.Slice {
		panic("redis: model is not a struct")
	}

	if minPK < 0 || maxPK < 0 {
		c.createSlice(nil)
		return nil
	}

	keysList, err := c.listKeys(minPK, maxPK, limit, isOrderAsc)
	if err != nil {
		return err
	}
	fields := c.getFields(nil, skippedFields)

	ret, err := c.redisCli.Pipelined(func(pipe redis.Pipeliner) error {
		for _, key := range keysList {
			pipe.HMGet(key, fields...)
		}
		return nil
	})
	if err != nil {
		return err
	}

	var (
		objVal, structVal, fieldVal reflect.Value
		objs                        []reflect.Value
		v                           []interface{}
		empty                       bool
		field                       *Field
	)
	for _, cmd := range ret {
		objVal = reflect.New(c.table.Type)
		v, err = cmd.(*redis.SliceCmd).Result()
		if err != nil {
			return err
		}

		// We need a struct so check if it's a Ptr.
		structVal = objVal
		if structVal.Kind() == reflect.Ptr {
			structVal = structVal.Elem()
		}

		empty = true
		for i, f := range fields {
			field = c.table.Fields[f]

			if v[i] == nil {
				if !field.HasDefault() {
					continue
				}
				fieldVal = field.Default()
			} else {
				fieldVal = reflect.ValueOf(field.Adapter.Load(v[i].(string)))
				empty = false
			}
			field.Value(structVal).Set(fieldVal)
		}
		if !empty {
			objs = append(objs, objVal)
		}
	}

	c.createSlice(objs)
	return nil
}

func (c *DBCtx) trimList(cmds []redis.Cmder) error {
	trimmedTTL := c.model.TrimmedTTL()
	if trimmedTTL <= 0 {
		return nil
	}

	keys, err := cmds[len(cmds)-2].(*redis.StringSliceCmd).Result()
	if err != nil {
		return err
	}
	_, err = c.redisCli.Pipelined(func(pipe redis.Pipeliner) error {
		for _, key := range keys {
			pipe.Expire(key, trimmedTTL)
		}
		return nil
	})
	return err
}

func (c *DBCtx) saveObject(pipe redis.Pipeliner, pk int, objectKey string, fields []string, saved bool, ttl time.Duration) bool {
	var (
		field    *Field
		v        reflect.Value
		s        string
		trimming bool
	)
	for _, f := range fields {
		field = c.table.Fields[f]
		v = field.Value(c.value)
		s = field.Adapter.Dump(v.Interface())

		if s != "" {
			pipe.HSet(objectKey, f, s)
		} else if saved {
			pipe.HDel(objectKey, f)
		}
	}
	if ttl > 0 {
		pipe.Expire(objectKey, ttl)
	}

	if !saved {
		// Save to list if not added already.
		listKey := c.getListKey()
		pipe.ZAdd(listKey, redis.Z{Score: float64(pk), Member: objectKey})
		if ttl > 0 {
			pipe.Expire(listKey, ttl)
		}

		listMaxSize := c.model.ListMaxSize()
		if listMaxSize > 0 && pk > listMaxSize {
			trimming = true
			trim := int64(-1 * (listMaxSize + 1))
			pipe.ZRange(listKey, 0, trim)
			pipe.ZRemRangeByRank(listKey, 0, trim)
		}
	}
	return trimming
}

// Save ...
func (c *DBCtx) Save(updateFields []string) error {
	if c.value.Kind() != reflect.Struct {
		panic("redis: model is not a struct")
	}
	ttl := c.model.TTL()
	pk := c.table.PK(c.value)
	saved := pk != 0

	// Set ID if not saved.
	if !saved {
		if len(updateFields) > 0 {
			panic("redis: updateFields cannot be specified for unsaved object")
		}
		seqKey := fmt.Sprintf("%s:seq", c.model.Key(c.args))
		v, err := c.redisCli.Incr(seqKey).Result()
		if err != nil {
			return err
		}
		if ttl > 0 {
			c.redisCli.Expire(seqKey, ttl*2)
		}
		i := int(v)
		c.table.SetPK(c.value, i)
		pk = i
	}

	// Save object.
	objectKey := c.getObjectKey(pk)
	fields := c.getFields(updateFields, nil)
	var trimming bool
	cmds, err := c.redisCli.Pipelined(func(pipe redis.Pipeliner) error {
		trimming = c.saveObject(pipe, pk, objectKey, fields, saved, ttl)
		return nil
	})
	if err != nil {
		return err
	}

	if trimming {
		return c.trimList(cmds)
	}
	return nil
}

// Delete ...
func (c *DBCtx) Delete() error {
	if c.value.Kind() != reflect.Struct {
		panic("redis: model is not a struct")
	}
	objectKey := c.getObjectKey(c.table.PK(c.value))
	listKey := c.getListKey()

	_, err := c.redisCli.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.Del(objectKey)
		pipe.ZRem(listKey, objectKey)
		return nil
	})
	return err
}

// Update ...
func (c *DBCtx) Update(pk int, updated, expected map[string]interface{}) error {
	objectKey := c.getObjectKey(pk)

	var watch []string
	if len(expected) > 0 {
		watch = append(watch, objectKey)
	}
	return util.RetryWithCritical(updateRetries, 0, func() (bool, error) {
		err := c.redisCli.Watch(func(tx *redis.Tx) error {
			var (
				cur string
				e   error
			)
			// Check expected first.
			for k, v := range expected {
				cur, e = tx.HGet(objectKey, k).Result()
				if e != nil {
					return e
				}
				if cur != c.table.Fields[k].Adapter.Dump(v) {
					return ErrExpectedMismatch
				}
			}

			// Process actual saving.
			_, e = tx.Pipelined(func(pipe redis.Pipeliner) error {
				var (
					field *Field
					ok    bool
					s     string
				)
				for f, v := range updated {
					field, ok = c.table.Fields[f]
					if !ok {
						panic(fmt.Sprintf("redis: undefined field %s", f))
					}
					s = field.Adapter.Dump(v)

					if s != "" {
						pipe.HSet(objectKey, f, s)
					} else {
						pipe.HDel(objectKey, f)
					}
				}
				ttl := c.model.TTL()
				if ttl > 0 {
					pipe.Expire(objectKey, ttl)
				}

				return nil
			})
			return e
		}, watch...)

		if err == redis.TxFailedErr {
			return false, err
		}
		return true, err
	})
}
