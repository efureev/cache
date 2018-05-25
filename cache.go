package cache

import (
	"sync"
	"time"
	"runtime"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type Cache struct {
	sync.RWMutex
	defaultExpiration time.Duration // продолжительность жизни кеша по-умолчанию
	cleanupInterval   time.Duration // интервал, через который запускается механизм очистки кеша
	items             map[string]Item
	cg                *cg
}

type Item struct {
	Value      interface{}
	Created    time.Time
	Expiration int64
}

/*
func (c *Cache) StartGC() {
	go c.GC()
}
*/
/*
func (c *Cache) GC() {

	for {
		<-time.After(c.cleanupInterval)

		if c.items == nil || len(c.items) == 0 {
			return
		}

		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}
	}
}
*/
func (c *Cache) expiredKeys() (keys []string) {
	for k, i := range c.items {
		if i.Expiration > 0 && time.Now().UnixNano() > i.Expiration {
			keys = append(keys, k)
		}
	}

	return keys
}

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

func (c *Cache) clearItems(keys []string) {
	for _, k := range keys {
		delete(c.items, k)
	}
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {

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

func (c *Cache) Get(key string) (interface{}, bool) {
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

func (c *Cache) Delete(key string) error {
	c.Lock()

	if _, found := c.items[key]; !found {
		return notFound
	}

	delete(c.items, key)

	c.Unlock()

	return nil
}

func (c *Cache) Count() int {
	c.RLock()
	n := len(c.items)
	c.RUnlock()
	return n
}

func (c *Cache) Flush() {
	c.Lock()
	c.items = map[string]Item{}
	c.Unlock()
}

func newCacheWithCG(defaultExpiration, cleanupInterval time.Duration, items map[string]Item) *Cache {
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

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return newCacheWithCG(defaultExpiration, cleanupInterval, make(map[string]Item))
}

func NewFrom(defaultExpiration, cleanupInterval time.Duration, items map[string]Item) *Cache {
	return newCacheWithCG(defaultExpiration, cleanupInterval, items)
}