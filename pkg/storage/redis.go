package storage

import (
	"github.com/go-redis/redis"

	"github.com/Syncano/orion/pkg/redisdb"
)

var (
	redisCli    *redis.Client
	redisDB     *redisdb.DB
	redisPubSub *PubSub
)

// InitRedis sets up Redis client.
func InitRedis(opts *redis.Options) {
	redisCli = redis.NewClient(opts)
	redisDB = redisdb.Init(redisCli)
	redisPubSub = NewPubSub(redisCli)
}

// Redis returns Redis client.
func Redis() *redis.Client {
	return redisCli
}

// RedisDB returns RedisDB client.
func RedisDB() *redisdb.DB {
	return redisDB
}

// RedisPubSub returns default Redis PubSub.
func RedisPubSub() *PubSub {
	return redisPubSub
}
