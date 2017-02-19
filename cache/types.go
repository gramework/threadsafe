package cache

import (
	"sync"

	"github.com/gramework/utils/nocopy"
)

// Instance represents a cache instance
type Instance struct {
	storage map[string]interface{}
	nocopy  nocopy.NoCopy
	lock    sync.RWMutex
}
