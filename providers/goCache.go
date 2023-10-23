package providers

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/wasilak/cachego/config"
)

// The GoCache type represents a cache with a specified time-to-live duration and optional tracing and
// context.
// @property Cache - The `Cache` property is a pointer to an instance of the `gocache.Cache` struct.
// This struct represents a cache that can be used to store and retrieve data efficiently.
// @property TTL - TTL stands for "Time to Live". It is a duration that specifies the amount of time
// for which an item in the cache should be considered valid before it expires and is removed from the
// cache.
// @property Tracer - The Tracer property is of type trace.Tracer. It is used for tracing and
// monitoring the cache operations.
// @property CTX - CTX is a context.Context object. It is used to carry request-scoped values across
// API boundaries and between processes. It allows cancellation signals and request-scoped values to
// propagate across API boundaries and between processes.
type GoCache struct {
	Cache  *gocache.Cache
	Config config.CacheGoConfig
}

func (c *GoCache) GetConfig() config.CacheGoConfig {
	return c.Config
}

// The `Init` function is initializing the cache by creating a new instance of `gocache.Cache` with the
// specified time-to-live duration (`TTL`). It also starts a new span using the provided tracer and
// context for tracing and monitoring purposes. Finally, it assigns the newly created cache to the
// `Cache` property of the `GoCache` struct.
func (c *GoCache) Init() error {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "Init")
	defer span.End()

	c.Cache = gocache.New(c.Config.TTL, c.Config.TTL)

	return nil
}

func (c *GoCache) Get(cacheKey string) ([]byte, bool, error) {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "Get")
	defer span.End()

	item, found := c.Cache.Get(cacheKey)

	if item == nil {
		var empty []byte
		return empty, found, nil
	}

	return item.([]byte), found, nil
}

// The `Set` function is used to store an item in the cache. It takes two parameters: `cacheKey`, which
// is a string representing the key for the item, and `item`, which is the actual item to be stored in
// the cache. The function uses the `cacheKey` and `item` parameters to set the value in the cache
// using the `gocache.Set` method. It also sets the time-to-live (TTL) for the item to the value
// specified in the `TTL` property of the `GoCache` struct. Finally, it returns an error if any
// occurred during the operation.
func (c *GoCache) Set(cacheKey string, item []byte) error {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "Set")
	defer span.End()

	c.Cache.Set(cacheKey, item, c.Config.TTL)

	return nil
}

// The `GetItemTTL` function is used to retrieve the remaining time-to-live (TTL) duration for a
// specific item in the cache. It takes a `cacheKey` parameter, which is a string representing the key
// of the item.
func (c *GoCache) GetItemTTL(cacheKey string) (time.Duration, bool, error) {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "GetItemTTL")
	defer span.End()

	_, expiration, found := c.Cache.GetWithExpiration(cacheKey)

	now := time.Now()
	difference := expiration.Sub(now)

	return difference, found, nil
}

// The `ExtendTTL` function is used to extend the time-to-live (TTL) duration of a specific item in the
// cache. It takes two parameters: `cacheKey`, which is a string representing the key of the item, and
// `item`, which is the updated value of the item.
func (c *GoCache) ExtendTTL(cacheKey string, item []byte) error {
	_, span := c.Config.Tracer.Start(c.Config.CTX, "ExtendTTL")
	defer span.End()

	c.Set(cacheKey, item)

	return nil
}
