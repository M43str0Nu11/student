package cache

import (
	"orders-demo/internal/model"
	"sync"
)

type Cache struct {
	mu     sync.RWMutex
	orders map[string]model.CachedOrder
}

func New() *Cache {
	return &Cache{orders: make(map[string]model.CachedOrder)}
}

func (c *Cache) Set(id string, order model.CachedOrder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[id] = order
}

func (c *Cache) Get(id string) (model.CachedOrder, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.orders[id]
	return order, ok
}

func (c *Cache) LoadAll(data map[string]model.CachedOrder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range data {
		c.orders[k] = v
	}
}
