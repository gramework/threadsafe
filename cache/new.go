package cache

import "sync"

// New Instance
func New() *Instance {
	return &Instance{
		storage: make(map[string]interface{}),
		lock:    sync.RWMutex{},
	}
}
