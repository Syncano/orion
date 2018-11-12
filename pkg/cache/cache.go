package cache

import (
	"reflect"
	"time"

	"github.com/go-redis/cache"
	"github.com/vmihailenco/msgpack"

	"github.com/Syncano/orion/pkg/settings"
	"github.com/Syncano/orion/pkg/util"
)

var (
	codec, codecLocal *cache.Codec
)

const (
	versionGraceDuration = 5 * time.Minute
)

// Init sets up a cache.
func Init(cli rediser) {
	codec = &cache.Codec{
		Redis: cli,

		Marshal:   msgpack.Marshal,
		Unmarshal: msgpack.Unmarshal,
	}
	codecLocal = &cache.Codec{
		Marshal:   msgpack.Marshal,
		Unmarshal: msgpack.Unmarshal,
	}
	codecLocal.UseLocalCache(50000, settings.Common.LocalCacheTimeout)
}

// Codec returns cache client.
func Codec() *cache.Codec {
	return codec
}

// CodecLocal returns cache client that uses local cache as well as remote.
func CodecLocal() *cache.Codec {
	return codecLocal
}

// Stats returns cache statistics.
func Stats() *cache.Stats {
	stats := codec.Stats()
	statsLocal := codec.Stats()
	statsLocal.Hits += stats.Hits
	statsLocal.Misses += stats.Misses
	return statsLocal
}

type cacheItem struct {
	Object  interface{}
	Version string
}

func (ci *cacheItem) validate(version string, validate func(interface{}) bool) bool {
	return version == ci.Version && (validate == nil || validate(ci.Object))
}

// VersionedCache ...
func VersionedCache(cacheKey string, lookup string, val interface{},
	versionKeyFunc func() string, compute func() (interface{}, error), validate func(interface{}) bool, expiration time.Duration) error {

	item := &cacheItem{Object: val}
	var version string

	// Get object and check version. First local and fallback to global cache.
	if codecLocal.Get(cacheKey, item) == nil {
		version = codec.Redis.Get(versionKeyFunc()).Val()
		if item.validate(version, validate) {
			return nil
		}
	}
	if codec.Get(cacheKey, item) == nil {
		if version == "" {
			version = codec.Redis.Get(versionKeyFunc()).Val()
		}
		if item.validate(version, validate) {
			return codecLocal.Set(&cache.Item{
				Key:        cacheKey,
				Object:     item,
				Expiration: settings.Common.LocalCacheTimeout,
			})
		}
	}

	// Compute and save object.
	object, err := compute()
	if err != nil {
		return err
	}

	if version == "" {
		version = codec.Redis.Get(versionKeyFunc()).Val()
	}

	// Set object through reflect.
	vref := reflect.ValueOf(val)
	oref := reflect.ValueOf(object)

	if oref.Kind() == reflect.Ptr {
		oref = oref.Elem()
	}
	vref.Elem().Set(oref)
	item.Object = val
	item.Version = version

	// Set cache values.
	codecLocal.Set(&cache.Item{ // nolint: errcheck
		Key:        cacheKey,
		Object:     item,
		Expiration: settings.Common.LocalCacheTimeout,
	})
	return codec.Set(&cache.Item{
		Key:        cacheKey,
		Object:     item,
		Expiration: expiration,
	})
}

// InvalidateVersion ...
func InvalidateVersion(versionKey string, expiration time.Duration) {
	codec.Redis.Set(
		versionKey,
		util.GenerateRandomString(4),
		expiration+versionGraceDuration, // Add grace period to avoid race condition.
	)
}
