package cache

import (
	"sync"

	"github.com/gramework/utils/nocopy"
)

// Instance represents a cache instance
type Instance struct {
	storage map[string]interface{}
	lock    sync.RWMutex

	nocopy nocopy.NoCopy
}
