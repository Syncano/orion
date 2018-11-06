package storage

import "github.com/go-redis/redis"

var (
	redisCli *redis.Client
)

// InitRedis sets up redis client.
func InitRedis(opts *redis.Options) {
	redisCli = redis.NewClient(opts)
}

// Redis returns redis client.
func Redis() *redis.Client {
	return redisCli
}
