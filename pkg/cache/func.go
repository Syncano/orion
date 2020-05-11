package cache

import (
	"fmt"

	"github.com/go-pg/pg/v9/orm"

	"github.com/Syncano/orion/pkg/util"
)

func (c *Cache) createFuncCacheKey(funcKey, versionKey, lookup string) string {
	return fmt.Sprintf("0:cache:f:%d:%s:%s:%x", c.opts.CacheVersion, funcKey, versionKey, util.Hash(lookup))
}

func (c *Cache) createFuncVersionCacheKey(funcKey, versionKey string) string {
	return fmt.Sprintf("0:cache:f:%d:%s:%s:version", c.opts.CacheVersion, funcKey, versionKey)
}

func (c *Cache) FuncCacheInvalidate(funcKey, versionKey string) error {
	versionKey = c.createFuncVersionCacheKey(funcKey, versionKey)
	return c.InvalidateVersion(versionKey, c.opts.CacheTimeout)
}

func (c *Cache) FuncCacheCommitInvalidate(db orm.DB, funcKey, versionKey string) {
	c.db.AddDBCommitHook(db, func() error {
		return c.FuncCacheInvalidate(funcKey, versionKey)
	})
}

func (c *Cache) FuncCache(funcKey, versionKey string, val interface{}, lookup string,
	compute func() (interface{}, error), validate func(interface{}) bool) error {
	funcKey = c.createFuncCacheKey(funcKey, versionKey, lookup)

	return c.VersionedCache(funcKey, lookup, val,
		func() string {
			return c.createFuncVersionCacheKey(funcKey, versionKey)
		},
		compute, validate, c.opts.CacheTimeout)
}

func (c *Cache) SimpleFuncCache(funcKey, versionKey string, val interface{}, lookup string,
	compute func() (interface{}, error)) error {
	return c.FuncCache(funcKey, versionKey, val, lookup, compute, nil)
}
