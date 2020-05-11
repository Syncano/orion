package storage

import (
	"github.com/go-redis/redis/v7"

	"github.com/Syncano/orion/pkg/redisdb"
)

type Redis struct {
	cli    *redis.Client
	db     *redisdb.DB
	pubsub *PubSub
}

// InitRedis sets up Redis client.
func NewRedis(opts *redis.Options) *Redis {
	redisCli := redis.NewClient(opts)

	return &Redis{
		cli:    redisCli,
		db:     redisdb.New(redisCli),
		pubsub: NewPubSub(redisCli),
	}
}

// Redis returns Redis client.
func (r *Redis) Client() *redis.Client {
	return r.cli
}

// RedisDB returns RedisDB client.
func (r *Redis) DB() *redisdb.DB {
	return r.db
}

// RedisPubSub returns default Redis PubSub.
func (r *Redis) PubSub() *PubSub {
	return r.pubsub
}
