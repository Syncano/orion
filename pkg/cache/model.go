package cache

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/pkg/storage"
	"github.com/Syncano/orion/pkg/util"
)

func (c *Cache) createModelCacheKey(schema, model, lookup string) string {
	return fmt.Sprintf("%s:cache:m:%d:%s:%x", schema, c.opts.CacheVersion, model, util.Hash(lookup))
}

func (c *Cache) createModelVersionCacheKey(schema, model string, pk interface{}) string {
	return fmt.Sprintf("%s:cache:m:%d:%s:%v:version", schema, c.opts.CacheVersion, model, pk)
}

func getSchemaKey(db orm.DB) string {
	schema, ok := db.Context().Value(storage.KeySchema).(string)
	if !ok {
		schema = "0"
	} else {
		schema = strings.SplitN(schema, "_", 2)[0]
	}

	return schema
}

func (c *Cache) ModelCacheInvalidate(db orm.DB, m interface{}) {
	c.db.AddDBCommitHook(db, func() error {
		table := orm.GetTable(reflect.TypeOf(m).Elem())
		tableName := string(table.FullName)
		schema := getSchemaKey(db)
		versionKey := c.createModelVersionCacheKey(schema, tableName, table.PKs[0].Value(reflect.ValueOf(m).Elem()).Interface())

		return c.InvalidateVersion(versionKey, c.opts.CacheTimeout)
	})
}

func (c *Cache) ModelCache(db orm.DB, keyModel, val interface{}, lookup string,
	compute func() (interface{}, error), validate func(interface{}) bool) error {
	table := orm.GetTable(reflect.TypeOf(keyModel).Elem())
	n := strings.Split(string(table.FullName), ".")
	tableName := n[len(n)-1]
	schema := getSchemaKey(db)
	modelKey := c.createModelCacheKey(schema, tableName, lookup)

	return c.VersionedCache(modelKey, lookup, val,
		func() string {
			return c.createModelVersionCacheKey(schema, tableName, table.PKs[0].Value(reflect.ValueOf(keyModel).Elem()))
		},
		compute, validate, c.opts.CacheTimeout)
}

func (c *Cache) SimpleModelCache(db orm.DB, m interface{}, lookup string, compute func() (interface{}, error)) error {
	return c.ModelCache(db, m, m, lookup, compute, nil)
}
