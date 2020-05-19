package cache

import (
	"reflect"
	"time"

	"github.com/go-redis/cache/v7"
	"github.com/go-redis/redis/v7"
	"github.com/imdario/mergo"
	"github.com/vmihailenco/msgpack/v4"

	"github.com/Syncano/orion/pkg/storage"
	"github.com/Syncano/orion/pkg/util"
)

const (
	versionGraceDuration = 5 * time.Minute
)

type Cache struct {
	codec, codecLocal *cache.Codec
	db                *storage.Database
	opts              *Options
}

type Options struct {
	LocalCacheTimeout time.Duration
	CacheTimeout      time.Duration
	CacheVersion      int
}

var DefaultOptions = &Options{
	CacheVersion:      1,
	CacheTimeout:      12 * time.Hour,
	LocalCacheTimeout: 1 * time.Hour,
}

// Init sets up a cache.
func New(r rediser, db *storage.Database, opts *Options) *Cache {
	if opts != nil {
		mergo.Merge(opts, DefaultOptions) // nolint - error not possible
	} else {
		opts = DefaultOptions
	}

	codec := &cache.Codec{
		Redis: r,

		Marshal:   msgpack.Marshal,
		Unmarshal: msgpack.Unmarshal,
	}
	codecLocal := &cache.Codec{
		Marshal:   msgpack.Marshal,
		Unmarshal: msgpack.Unmarshal,
	}
	codecLocal.UseLocalCache(50000, opts.LocalCacheTimeout)

	return &Cache{
		codec:      codec,
		codecLocal: codecLocal,
		db:         db,
		opts:       opts,
	}
}

// Codec returns cache client.
func (c *Cache) Codec() *cache.Codec {
	return c.codec
}

// CodecLocal returns cache client that uses local cache as well as remote.
func (c *Cache) CodecLocal() *cache.Codec {
	return c.codecLocal
}

// Stats returns cache statistics.
func (c *Cache) Stats() *cache.Stats {
	stats := c.codec.Stats()
	statsLocal := c.codec.Stats()
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

func (c *Cache) VersionedCache(cacheKey, lookup string, val interface{},
	versionKeyFunc func() string, compute func() (interface{}, error), validate func(interface{}) bool, expiration time.Duration) error {
	item := &cacheItem{Object: val}

	var (
		version string
		err     error
	)

	// Get object and check version. First local and fallback to global cache.
	if c.codecLocal.Get(cacheKey, item) == nil {
		version, err = c.codec.Redis.Get(versionKeyFunc()).Result()
		if err != nil && err != redis.Nil {
			return err
		}

		if item.validate(version, validate) {
			return nil
		}
	}

	if c.codec.Get(cacheKey, item) == nil {
		if version == "" {
			version, err = c.codec.Redis.Get(versionKeyFunc()).Result()
			if err != nil && err != redis.Nil {
				return err
			}
		}

		if item.validate(version, validate) {
			return c.codecLocal.Set(&cache.Item{
				Key:        cacheKey,
				Object:     item,
				Expiration: c.opts.LocalCacheTimeout,
			})
		}
	}

	// Compute and save object.
	object, err := compute()
	if err != nil {
		return err
	}

	if version == "" {
		version, err = c.codec.Redis.Get(versionKeyFunc()).Result()
		if err != nil && err != redis.Nil {
			return err
		}
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
	c.codecLocal.Set(&cache.Item{ // nolint: errcheck
		Key:        cacheKey,
		Object:     item,
		Expiration: c.opts.LocalCacheTimeout,
	})

	return c.codec.Set(&cache.Item{
		Key:        cacheKey,
		Object:     item,
		Expiration: expiration,
	})
}

func (c *Cache) InvalidateVersion(versionKey string, expiration time.Duration) error {
	return c.codec.Redis.Set(
		versionKey,
		util.GenerateRandomString(4),
		expiration+versionGraceDuration, // Add grace period to avoid race condition.
	).Err()
}
