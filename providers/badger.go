package providers

import (
	"context"
	"encoding/json"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"go.opentelemetry.io/otel/trace"
)

// The BadgerCache type represents a cache with a Badger database, a time-to-live duration, a tracer, a
// context, and a path.
// @property Cache - The `Cache` property is a pointer to a `badger.DB` object. `badger.DB` is a
// key-value store database that provides fast and efficient storage and retrieval of data. It is
// commonly used for caching purposes in Go applications.
// @property TTL - TTL stands for "Time to Live" and it represents the duration for which an item in
// the cache is considered valid before it expires and needs to be refreshed or reloaded.
// @property Tracer - The Tracer property is a variable of type trace.Tracer. It is used for tracing
// and monitoring purposes, allowing you to track the execution of your code and identify any
// performance issues or bottlenecks.
// @property CTX - CTX is a context.Context object that is used for managing the context of operations
// performed by the BadgerCache. It allows for cancellation, timeouts, and passing values across API
// boundaries.
// @property {string} Path - The `Path` property is a string that represents the file path where the
// BadgerCache database is stored.
type BadgerCache struct {
	Cache  *badger.DB
	TTL    time.Duration
	Tracer trace.Tracer
	CTX    context.Context
	Path   string
}

// The BadgerItem type represents an item with a time-to-live (TTL) and content.
// @property TTL - TTL stands for "Time to Live" and it represents the expiration time for the item. It
// is a time.Time type, which means it stores a specific point in time.
// @property Content - The "Content" property is an interface type, which means it can hold values of
// any type. It can be used to store any kind of data, such as strings, numbers, booleans, or even
// custom types.
type BadgerItem struct {
	TTL     time.Time
	Content interface{}
}

// The `Init` function is used to initialize the BadgerCache. It opens a connection to the Badger
// database using the provided path and sets the Cache field of the BadgerCache struct to the opened
// database. If any error occurs during the initialization process, it is returned.
func (c *BadgerCache) Init() error {
	_, span := c.Tracer.Start(c.CTX, "Init")
	defer span.End()

	opts := badger.DefaultOptions(c.Path)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return err
	}

	c.Cache = db

	return nil
}

// The `Get` function is used to retrieve an item from the cache based on a given cache key. It takes a
// cache key as input and returns three values: the content of the item (as an interface{}), a boolean
// indicating if the item exists in the cache, and an error if any occurred.
func (c *BadgerCache) Get(cacheKey string) (interface{}, bool, error) {
	_, span := c.Tracer.Start(c.CTX, "Get")
	defer span.End()

	item, err := c.retrieveFromCache(cacheKey)
	if err != nil {
		return nil, false, err
	}

	now := time.Now()

	if item.TTL.Unix() <= now.Unix() {
		c.delete(cacheKey)
		return item, false, nil
	}

	return item.Content, true, nil
}

// The `Set` function is used to store an item in the cache. It takes a cache key and an item as input.
// The function serializes the item into bytes using JSON encoding and creates a `BadgerItem` struct
// with the serialized item and a TTL (time to live) value. It then starts a transaction, sets the
// cache key-value pair in the transaction, and commits the transaction to persist the changes in the
// cache. If any error occurs during the process, it is returned.
func (c *BadgerCache) Set(cacheKey string, item interface{}) error {
	_, span := c.Tracer.Start(c.CTX, "Set")
	defer span.End()

	ttl := time.Now().Add(c.TTL)

	badgerItem := BadgerItem{
		Content: item,
		TTL:     ttl,
	}

	// Serialize the item to bytes
	itemBytes, err := json.Marshal(badgerItem)
	if err != nil {
		return err
	}

	// Start a transaction
	txn := c.Cache.NewTransaction(true)
	defer txn.Discard()

	// Set the cache key-value pair
	if err := txn.Set([]byte(cacheKey), itemBytes); err != nil {
		return err
	}

	// Commit the transaction
	if err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

// The `GetItemTTL` function is used to retrieve the remaining time to live (TTL) of an item in the
// cache. It takes a cache key as input and returns the remaining TTL duration, a boolean indicating if
// the item exists in the cache, and an error if any occurred.
func (c *BadgerCache) GetItemTTL(cacheKey string) (time.Duration, bool, error) {
	_, span := c.Tracer.Start(c.CTX, "GetItemTTL")
	defer span.End()

	var itemTTL time.Duration

	item, err := c.retrieveFromCache(cacheKey)
	if err != nil {
		return itemTTL, false, err
	}

	now := time.Now()
	difference := item.TTL.Sub(now)

	return difference, true, nil
}

// The `ExtendTTL` function is used to extend the time to live (TTL) of an item in the cache. It takes
// a cache key and an item as input. The function calls the `Set` function to update the item in the
// cache with a new TTL. This effectively extends the lifespan of the item in the cache.
func (c *BadgerCache) ExtendTTL(cacheKey string, item interface{}) error {
	_, span := c.Tracer.Start(c.CTX, "ExtendTTL")
	defer span.End()

	c.Set(cacheKey, item)

	return nil
}

// The `retrieveFromCache` function is used to retrieve an item from the cache based on a given cache
// key. It takes a cache key as input and returns a `BadgerItem` struct and an error.
func (c *BadgerCache) retrieveFromCache(cacheKey string) (BadgerItem, error) {
	txn := c.Cache.NewTransaction(false)
	defer txn.Discard()

	var itemValue BadgerItem

	item, err := txn.Get([]byte(cacheKey))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return itemValue, nil
		}
		return itemValue, err
	}

	err = item.Value(func(val []byte) error {
		// Deserialize the value into the appropriate type
		if err := json.Unmarshal(val, &itemValue); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return itemValue, err
	}

	return itemValue, nil
}

// The `delete` function is used to delete an item from the cache based on a given cache key. It takes
// a cache key as input and returns an error if any occurred.
func (c *BadgerCache) delete(cacheKey string) error {
	_, span := c.Tracer.Start(c.CTX, "Delete")
	defer span.End()

	// Start a transaction
	txn := c.Cache.NewTransaction(true)
	defer txn.Discard()

	// Delete the item by key
	if err := txn.Delete([]byte(cacheKey)); err != nil {
		return err
	}

	// Commit the transaction
	if err := txn.Commit(); err != nil {
		return err
	}

	return nil
}
