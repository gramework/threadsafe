package store

import (
	"github.com/gramework/threadsafe/hashmap"
	"github.com/gramework/utils/nocopy"
)

// Store itself
type Store struct {
	store hashmap.Map

	nocopy nocopy.NoCopy
}

// Put or replace a key
func (s *Store) Put(key string, v interface{}) {
	s.store.Put(key, v)
}

// Get a key from the storage
func (s *Store) Get(key string) (v interface{}, ok bool) {
	v, ok = s.store.GetPtrOk(key)
	return
}
