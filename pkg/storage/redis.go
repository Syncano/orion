package storage

import (
	"github.com/go-redis/redis"

	"github.com/Syncano/orion/pkg/redisdb"
)

var (
	redisCli *redis.Client
	redisDB  *redisdb.DB
)

// InitRedis sets up Redis client.
func InitRedis(opts *redis.Options) {
	redisCli = redis.NewClient(opts)
	redisDB = redisdb.Init(redisCli)
}

// Redis returns Redis client.
func Redis() *redis.Client {
	return redisCli
}

// RedisDB returns RedisDB client.
func RedisDB() *redisdb.DB {
	return redisDB
}
