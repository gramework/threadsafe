package cache

func (c *Cache) Put(key string, value interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.storage[key] = value
	return nil
}
