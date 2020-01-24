package redisdb

import (
	"fmt"
	"reflect"

	"github.com/go-redis/redis"
)

var (
	modelerType = reflect.TypeOf((*Modeler)(nil)).Elem()
)

// DB represents Redis DB object.
type DB struct {
	redisCli *redis.Client
}

// Init sets up Redis DB.
func Init(cli *redis.Client) *DB {
	return &DB{
		redisCli: cli,
	}
}

// Model returns ctx for specified model and args.
func (d *DB) Model(m interface{}, args map[string]interface{}) *DBCtx {
	v := reflect.ValueOf(m)
	if !v.IsValid() {
		panic("redis: Model(nil)")
	}

	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("pg: Model(non-pointer %T)", m))
	}

	var typ reflect.Type

	if v.IsNil() {
		panic("redis: Model(nil)")
	}

	typ = v.Type()
	v = v.Elem()

	switch v.Kind() {
	case reflect.Slice:
		typ = v.Type().Elem()
	case reflect.Struct:
	default:
		panic(fmt.Sprintf("redis: Model(unsupported %s)", v.Type()))
	}

	if !typ.Implements(modelerType) {
		panic(fmt.Sprintf("redis: Model %s does not implement Modeler", v.Type()))
	}

	modeler := reflect.Zero(typ).Interface().(Modeler)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return &DBCtx{
		DB:   d,
		args: args,

		model: modeler,
		table: GetTable(typ),
		value: v,
	}
}
