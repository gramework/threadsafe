package cache

import "sync"

func New() *Cache {
	return &Cache{
		storage: make(map[string]interface{}),
		lock:    sync.RWMutex{},
	}
}
