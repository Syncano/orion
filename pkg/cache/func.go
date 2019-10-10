package cache

import (
	"fmt"

	"github.com/go-pg/pg/orm"

	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/storage"
	"github.com/Syncano/orion/pkg/util"
)

func createFuncCacheKey(funcKey, versionKey, lookup string) string {
	return fmt.Sprintf("0:cache:f:%d:%s:%s:%x", settings.Common.CacheVersion, funcKey, versionKey, util.Hash(lookup))
}

func createFuncVersionCacheKey(funcKey, versionKey string) string {
	return fmt.Sprintf("0:cache:f:%d:%s:%s:version", settings.Common.CacheVersion, funcKey, versionKey)
}

// FuncCacheInvalidate ...
func FuncCacheInvalidate(funcKey, versionKey string) error {
	versionKey = createFuncVersionCacheKey(funcKey, versionKey)
	return InvalidateVersion(versionKey, settings.Common.CacheTimeout)
}

// FuncCacheCommitInvalidate ...
func FuncCacheCommitInvalidate(db orm.DB, funcKey, versionKey string) {
	storage.AddDBCommitHook(db, func() error {
		return FuncCacheInvalidate(funcKey, versionKey)
	})
}

// FuncCache ...
func FuncCache(funcKey, versionKey string, val interface{}, lookup string,
	compute func() (interface{}, error), validate func(interface{}) bool) error {

	funcKey = createFuncCacheKey(funcKey, versionKey, lookup)

	return VersionedCache(funcKey, lookup, val,
		func() string {
			return createFuncVersionCacheKey(funcKey, versionKey)
		},
		compute, validate, settings.Common.CacheTimeout)
}

// SimpleFuncCache ...
func SimpleFuncCache(funcKey, versionKey string, val interface{}, lookup string,
	compute func() (interface{}, error)) error {
	return FuncCache(funcKey, versionKey, val, lookup, compute, nil)
}
