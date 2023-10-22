package cachego

import (
	"context"
	"log"
	"time"

	"dario.cat/mergo"
	"github.com/wasilak/cachego/providers"
	"go.opentelemetry.io/otel"
)

// The above code defines a CacheInterface type in Go that represents a cache with methods for
// initialization, getting and setting items, and managing item time-to-live (TTL).
// @property {error} Init - The Init method is used to initialize the cache. It can be used to set up
// any necessary data structures or connections required for caching.
// @property Get - The Get method retrieves an item from the cache based on the provided cache key. It
// returns the item, a boolean indicating whether the item was found in the cache, and an error if any
// occurred.
// @property {error} Set - The Set method is used to store an item in the cache. It takes a cacheKey
// string as the identifier for the item and the item itself as the value to be stored in the cache.
// @property GetItemTTL - GetItemTTL is a method that retrieves the time-to-live (TTL) value for a
// specific item in the cache. The TTL represents the amount of time that the item will remain in the
// cache before it expires and is automatically removed. The method returns the TTL value as a
// time.Duration,
// @property {error} ExtendTTL - ExtendTTL is a method that extends the time to live (TTL) of a cached
// item. It takes a cacheKey string and an item interface as parameters. The cacheKey is used to
// identify the cached item, and the item is the updated value that will be stored in the cache.
type CacheInterface interface {
	Init() error
	Get(cacheKey string) (interface{}, bool, error)
	Set(cacheKey string, item interface{}) error
	GetItemTTL(cacheKey string) (time.Duration, bool, error)
	ExtendTTL(cacheKey string, item interface{}) error
}

// The `CacheGoConfig` type represents the configuration for a cache in Go, including the type,
// expiration, Redis host, Redis database, and path.
// @property {string} Type - The "Type" property is used to specify the type of cache to be used. It
// can be set to values like "memory", "redis", or "file" depending on the desired cache
// implementation.
// @property {string} Expiration - The "Expiration" property in the CacheGoConfig struct represents the
// duration for which the cache entries will be considered valid before they expire. It is typically
// specified as a string in a time format, such as "1h" for 1 hour, "30m" for 30 minutes, or
// @property {string} RedisHost - The RedisHost property is used to specify the host address of the
// Redis server that the cache will connect to.
// @property {int} RedisDB - RedisDB is an integer property that represents the database number to be
// used for caching in Redis.
// @property {string} Path - The `Path` property is a string that represents the file path where the
// cache data will be stored.
type CacheGoConfig struct {
	Type       string
	Expiration string
	RedisHost  string
	RedisDB    int
	Path       string
}

// The `var defaultConfig = CacheGoConfig{...}` statement is initializing a variable named
// `defaultConfig` with a value of type `CacheGoConfig`. It is setting the properties of the
// `CacheGoConfig` struct with default values.
var defaultConfig = CacheGoConfig{
	Type:       "memory",
	Expiration: "10m",
	RedisHost:  "127.0.0.1:6379",
	RedisDB:    0,
	Path:       "/tmp/cachego",
}

// The line `var CacheInstance CacheInterface` is declaring a variable named `CacheInstance` of type
// `CacheInterface`. This variable will be used to store an instance of a cache that implements the
// `CacheInterface` interface.
var CacheInstance CacheInterface

// The function `CacheInit` initializes and returns a cache instance based on the provided
// configuration.
func CacheInit(ctx context.Context, config CacheGoConfig) (CacheInterface, error) {
	tracer := otel.Tracer("Cache")
	_, span := tracer.Start(ctx, "CacheInit")
	defer span.End()

	err := mergo.Merge(&defaultConfig, config, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	ttl, err := time.ParseDuration(defaultConfig.Expiration)
	if err != nil {
		return nil, err
	}

	switch defaultConfig.Type {
	case "memory":
		{
			CacheInstance = &providers.GoCache{
				TTL:    ttl,
				Tracer: otel.Tracer("GoCache"),
				CTX:    ctx,
			}
		}

	case "file":
	case "badger":
		{
			CacheInstance = &providers.BadgerCache{
				TTL:    ttl,
				Tracer: otel.Tracer("FileCache"),
				CTX:    ctx,
				Path:   defaultConfig.Path,
			}
		}

	case "redis":
		{
			CacheInstance = &providers.RedisCache{
				Address: defaultConfig.RedisHost,
				DB:      defaultConfig.RedisDB,
				TTL:     ttl,
				Tracer:  otel.Tracer("RedisCache"),
				CTX:     ctx,
			}
		}

	default:
		{
			log.Fatal("No cache type selected or cache type is invalid")
		}

	}

	CacheInstance.Init()

	return CacheInstance, nil
}
