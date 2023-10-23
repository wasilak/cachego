package config

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
)

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
	TTL        time.Duration
	Tracer     trace.Tracer
	CTX        context.Context
}

// The `var defaultConfig = CacheGoConfig{...}` statement is initializing a variable named
// `defaultConfig` with a value of type `CacheGoConfig`. It is setting the properties of the
// `CacheGoConfig` struct with default values.
var DefaultConfig = CacheGoConfig{
	CTX:        context.Background(),
	Type:       "memory",
	Expiration: "10m",
	RedisHost:  "127.0.0.1:6379",
	RedisDB:    0,
	Path:       "/tmp/cachego",
}
