package cache

import "errors"

func (c *Cache) Get(key string) (interface{}, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if v, ok := c.storage[key]; ok {
		return v, nil
	}
	return nil, errors.New(NotFound)
}
