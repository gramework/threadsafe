package cache

// Get a key from the cache
func (c *Instance) Get(key string) (interface{}, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if v, ok := c.storage[key]; ok {
		return v, nil
	}
	return nil, ErrNotFound
}
