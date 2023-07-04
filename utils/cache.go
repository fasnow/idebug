package utils

import (
	"sync"
	"time"
)

type Cache struct {
	data     map[string]interface{}
	expire   map[string]time.Time
	mutex    sync.RWMutex
	interval time.Duration
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		data:     make(map[string]interface{}),
		expire:   make(map[string]time.Time),
		interval: interval,
	}
	go c.startCleanup()
	return c
}

func (c *Cache) Set(key string, value interface{}, expire time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = value
	c.expire[key] = time.Now().Add(expire)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, ok := c.data[key]
	return value, ok
}

func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.interval)
	for range ticker.C {
		c.mutex.Lock()
		for key, expireTime := range c.expire {
			if time.Now().After(expireTime) {
				delete(c.data, key)
				delete(c.expire, key)
			}
		}
		c.mutex.Unlock()
	}
}
