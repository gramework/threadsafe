package cache

// Put the value in a key
func (c *Instance) Put(key string, value interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.storage[key] = value
	return nil
}
