package store

import (
	"sync"

	"github.com/gramework/utils/nocopy"
)

// Store itself
type Store struct {
	store map[string]interface{}
	*sync.RWMutex

	nocopy nocopy.NoCopy
}

// Put or replace a key
func (s *Store) Put(key string, v interface{}) {
	s.Lock()
	s.store[key] = v
	s.Unlock()
}

// Get a key from the storage
func (s *Store) Get(key string) (v interface{}, ok bool) {
	s.RLock()
	v, ok = s.store[key]
	s.RUnlock()
	return
}
