package cachego

import (
	"context"
	"log"
	"time"

	"dario.cat/mergo"
	"github.com/wasilak/cachego/config"
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
	Get(cacheKey string) (string, bool, error)
	GetConfig() config.CacheGoConfig
	Set(cacheKey string, item []byte) error
	GetItemTTL(cacheKey string) (time.Duration, bool, error)
	ExtendTTL(cacheKey string, item []byte) error
}

// The line `var CacheInstance CacheInterface` is declaring a variable named `CacheInstance` of type
// `CacheInterface`. This variable will be used to store an instance of a cache that implements the
// `CacheInterface` interface.
var CacheInstance CacheInterface

// The function `CacheInit` initializes and returns a cache instance based on the provided
// configuration.
func CacheInit(ctx context.Context, cacheConfig config.CacheGoConfig) (CacheInterface, error) {
	tracer := otel.Tracer("Cache")
	_, span := tracer.Start(ctx, "CacheInit")
	defer span.End()

	err := mergo.Merge(&config.DefaultConfig, cacheConfig, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	ttl, err := time.ParseDuration(config.DefaultConfig.Expiration)
	if err != nil {
		return nil, err
	}

	config.DefaultConfig.CTX = ctx
	config.DefaultConfig.TTL = ttl

	switch config.DefaultConfig.Type {
	case "memory":
		{
			config.DefaultConfig.Tracer = otel.Tracer("GoCache")
			CacheInstance = &providers.GoCache{
				Config: config.DefaultConfig,
			}
		}

	case "file", "badger":
		{
			config.DefaultConfig.Tracer = otel.Tracer("FileCache")
			CacheInstance = &providers.BadgerCache{
				Path:   config.DefaultConfig.Path,
				Config: config.DefaultConfig,
			}
		}

	case "redis":
		{
			config.DefaultConfig.Tracer = otel.Tracer("RedisCache")
			CacheInstance = &providers.RedisCache{
				Address: config.DefaultConfig.RedisHost,
				DB:      config.DefaultConfig.RedisDB,
				Config:  config.DefaultConfig,
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
