// Package datastore implements structured data storage for wash server functionality.
package datastore

import (
	"regexp"
	"sync"
	"time"

	// TODO: Once https://github.com/patrickmn/go-cache/pull/75
	// is merged, go back to importing the main go-cache repo.
	cache "github.com/ekinanp/go-cache"
	"github.com/hashicorp/vault/helper/locksutil"
	log "github.com/sirupsen/logrus"
)

// Cache is an interface for a cache.
type Cache interface {
	GetOrUpdate(category, key string, ttl time.Duration, resetTTLOnHit bool, generateValue func() (interface{}, error)) (interface{}, error)
	Flush()
	Delete(matcher *regexp.Regexp) []string
}

// MemCache is an in-memory cache. It supports concurrent get/set, as well as the ability
// to get-or-update cached data in a single transaction to avoid redundant update activity.
type MemCache struct {
	instance    *cache.Cache
	locks       sync.Map
	hasEviction bool
}

// NewMemCache creates a new MemCache object
func NewMemCache() *MemCache {
	// The TTLs will be passed-in individually in the GetOrUpdate
	// method so we don't need to specify a default expiration
	cache := cache.New(cache.NoExpiration, 1*time.Minute)
	return &MemCache{
		instance:    cache,
		hasEviction: false,
	}
}

// NewMemCacheWithEvicted creates a new MemCache object that calls the provided eviction function
// on each object as it's evicted to facilitate cleanup.
func NewMemCacheWithEvicted(f func(string, interface{})) *MemCache {
	cache := NewMemCache()
	cache.instance.OnEvicted(f)
	cache.hasEviction = true
	return cache
}

// LockForKey retrieve the lock used for a specific category/key pair.
func (cache *MemCache) lockForKey(category, key string) *locksutil.LockEntry {
	// If a lockset is present for the category, use it. Otherwise create one and add it.
	obj, ok := cache.locks.Load(category)
	if !ok {
		obj, _ = cache.locks.LoadOrStore(category, locksutil.CreateLocks())
	}
	return locksutil.LockForKey(obj.([]*locksutil.LockEntry), key)
}

// GetOrUpdate attempts to retrieve the value stored at the given key.
// If the value does not exist, then it generates the value using
// the generateValue function and stores it with the specified ttl.
// If resetTTLOnHit is true, will reset the cache expiration for the entry.
func (cache *MemCache) GetOrUpdate(category, key string, ttl time.Duration, resetTTLOnHit bool, generateValue func() (interface{}, error)) (interface{}, error) {
	l := cache.lockForKey(category, key)
	l.Lock()
	defer l.Unlock()

	// From here on key is a composition of category and key so we can maintain
	// a single cache.
	key = category + "::" + key
	value, found := cache.instance.Get(key)
	if found {
		log.Tracef("Cache hit on %v", key)
		if resetTTLOnHit {
			// Update last-access time
			cache.instance.Set(key, value, ttl)
		}
		if err, ok := value.(error); ok {
			return nil, err
		}
		return value, nil
	}

	// Cache misses should be rarer, so print them as debug messages.
	log.Debugf("Cache miss on %v", key)
	value, err := generateValue()
	// Cache error responses as well. These are often authentication or availability failures
	// and we don't want to continually query the API on failures.
	if err != nil {
		cache.instance.Set(key, err, ttl)
		return nil, err
	}

	cache.instance.Set(key, value, ttl)
	return value, nil
}

// Flush deletes all items from the cache. Also resets cache capacity.
// This operation is significantly slower when cache was created with NewMemCacheWithEvicted.
func (cache *MemCache) Flush() {
	if cache.hasEviction {
		// Flush doesn't trigger the eviction callback. If we've registered one, ensure it's
		// triggered for all keys being removed. First delete all valid entries, then delete
		// expired entries (the reverse would be incorrect, as entries might expire after
		// calling DeleteExpired but before calling Items).
		for k := range cache.instance.Items() {
			cache.instance.Delete(k)
		}
		cache.instance.DeleteExpired()
	}
	cache.instance.Flush()
}

// Delete removes entries from the cache that match the provided regexp.
func (cache *MemCache) Delete(matcher *regexp.Regexp) []string {
	log.Debugf("Deleting matches for %v", matcher)
	items := cache.instance.Items()
	deleted := make([]string, 0, len(items))
	for k := range items {
		if matcher.MatchString(k) {
			log.Debugf("Deleting cache entry %v", k)
			cache.instance.Delete(k)
			deleted = append(deleted, k)
		} else {
			log.Debugf("Skipping %v", k)
		}
	}
	return deleted
}
