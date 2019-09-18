package cache

import (
	"runtime"
	"sync"
	"time"
)

const (
	// NoExpiration - without Expiration
	NoExpiration time.Duration = -1

	// DefaultExpiration - 0
	DefaultExpiration time.Duration = 0
)

// Cache struct
type Cache struct {
	sync.RWMutex
	defaultExpiration time.Duration // продолжительность жизни кеша по-умолчанию
	cleanupInterval   time.Duration // интервал, через который запускается механизм очистки кеша
	items             map[interface{}]Item
	cg                *cg
}

// Item struct
type Item struct {
	Value      interface{}
	Created    time.Time
	Expiration int64
}

func (c *Cache) expiredKeys() (keys []interface{}) {
	for k, i := range c.items {
		if i.Expiration > 0 && time.Now().UnixNano() > i.Expiration {
			keys = append(keys, k)
		}
	}

	return keys
}

// DeleteExpired delete expired items
func (c *Cache) DeleteExpired() {

	c.Lock()
	defer c.Unlock()

	if c.items == nil || len(c.items) == 0 {
		return
	}

	if keys := c.expiredKeys(); len(keys) != 0 {
		c.clearItems(keys)
	}
}

func (c *Cache) clearItems(keys []interface{}) {
	for _, k := range keys {
		delete(c.items, k)
	}
}

// Set item to list by key
func (c *Cache) Set(key interface{}, value interface{}, duration time.Duration) {

	var expiration int64

	// Если продолжительность жизни равна 0 - используется значение по-умолчанию
	if duration == 0 {
		duration = c.defaultExpiration
	}

	// Устанавливаем время истечения кеша
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.Lock()

	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
		Created:    time.Now(),
	}

	c.Unlock()
}

// Get item by key
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	item, found := c.items[key]

	if !found {
		return nil, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}

	return item.Value, true
}

// Delete item by key
func (c *Cache) Delete(key interface{}) error {
	c.Lock()

	if _, found := c.items[key]; !found {
		return errNotFound
	}

	delete(c.items, key)

	c.Unlock()

	return nil
}

// Count items
func (c *Cache) Count() int {
	c.RLock()
	n := len(c.items)
	c.RUnlock()
	return n
}

// Flush clear items
func (c *Cache) Flush() {
	c.Lock()
	c.items = map[interface{}]Item{}
	c.Unlock()
}

func newCacheWithCG(defaultExpiration, cleanupInterval time.Duration, items map[interface{}]Item) *Cache {
	cache := Cache{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}

	if cleanupInterval > 0 {
		runCG(&cache, cleanupInterval)
		runtime.SetFinalizer(&cache, stopCG)
	}
	return &cache
}

// New create new instance
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return newCacheWithCG(defaultExpiration, cleanupInterval, make(map[interface{}]Item))
}

// NewFrom create new instance from item list
func NewFrom(defaultExpiration, cleanupInterval time.Duration, items map[interface{}]Item) *Cache {
	return newCacheWithCG(defaultExpiration, cleanupInterval, items)
}
