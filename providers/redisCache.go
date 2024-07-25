package providers

import (
	"time"

	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/wasilak/cachego/config"
)

// The RedisCache type represents a Redis cache with properties such as the cache client, time-to-live
// duration, address, database number, tracer, and context.
// @property Cache - The `Cache` property is a pointer to a Redis client. Redis is an open-source
// in-memory data structure store that can be used as a cache or a database. The Redis client allows
// you to interact with a Redis server and perform operations such as storing and retrieving data.
// @property TTL - TTL stands for "Time to Live" and it represents the duration for which a cache entry
// will be considered valid before it expires and is automatically removed from the cache.
// @property {string} Address - The `Address` property is a string that represents the address of the
// Redis server. It specifies the host and port where the Redis server is running.
// @property {int} DB - The `DB` property in the `RedisCache` struct represents the database number to
// be used for the Redis cache. Redis supports multiple databases, and each database is identified by a
// numeric index. By specifying the `DB` property, you can choose which database to use for storing the
// cache data.
// @property Tracer - The Tracer property is of type trace.Tracer. It is used for distributed tracing,
// which allows you to track and monitor requests as they flow through different services in a
// distributed system.
// @property CTX - CTX is a context.Context object that is used for managing the context of the
// RedisCache operations. It allows for cancellation, timeouts, and passing values across API
// boundaries.
type RedisCache struct {
	Cache   *redis.Client
	Address string
	DB      int
	Config  config.Config
}

func (c *RedisCache) GetConfig() config.Config {
	return c.Config
}

// The `Init` function is a method of the `RedisCache` struct. It initializes the Redis cache by
// creating a new Redis client and setting it to the `Cache` property of the `RedisCache` struct. The
// Redis client is created with the provided address and database number. The function returns an error
// if there is any issue initializing the Redis cache.
func (c *RedisCache) Init() error {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "Init")
	defer span.End()

	c.Cache = redis.NewClient(&redis.Options{
		Addr: c.Address,
		DB:   c.DB,
	})

	return nil
}

// The `Get` function is a method of the `RedisCache` struct. It is used to retrieve an item from the
// Redis cache based on the provided cache key.
func (c *RedisCache) Get(cacheKey string) ([]byte, bool, error) {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "Get")
	defer span.End()

	item, err := c.Cache.Get(c.Config.CTX, cacheKey).Bytes()

	switch {
	case err == redis.Nil:
		slog.Info("key does not exist", "key", cacheKey)
		return item, false, nil
	case err != nil:
		return item, false, err
	}

	if err != nil || len(item) == 0 {
		slog.ErrorContext(c.Config.CTX, "Error", slog.Any("message", err))
		return item, false, err
	}

	return item, true, nil
}

// The `Set` function is a method of the `RedisCache` struct. It is used to store an item in the Redis
// cache with the provided cache key.
func (c *RedisCache) Set(cacheKey string, item []byte) error {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "Set")
	defer span.End()

	return c.Cache.Set(c.Config.CTX, cacheKey, item, c.Config.TTL).Err()
}

// The `GetItemTTL` function is a method of the `RedisCache` struct. It is used to retrieve the
// remaining time-to-live (TTL) duration of an item in the Redis cache based on the provided cache key.
func (c *RedisCache) GetItemTTL(cacheKey string) (time.Duration, bool, error) {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "GetItemTTL")
	defer span.End()

	item, err := c.Cache.TTL(c.Config.CTX, cacheKey).Result()
	if err != nil {
		slog.ErrorContext(c.Config.CTX, "Error", slog.Any("message", err))
		return item, false, err
	}

	return item, true, nil
}

// The `ExtendTTL` function is a method of the `RedisCache` struct. It is used to extend the
// time-to-live (TTL) duration of an item in the Redis cache based on the provided cache key.
func (c *RedisCache) ExtendTTL(cacheKey string, item []byte) error {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "ExtendTTL")
	defer span.End()

	return c.Cache.Expire(c.Config.CTX, cacheKey, c.Config.TTL).Err()
}
