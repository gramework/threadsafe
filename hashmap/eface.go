package hashmap

import (
	"unsafe" // #nosec

	"github.com/gramework/runtimer"
)

type emptyInterface struct {
	typ  *runtimer.Type
	word unsafe.Pointer
}

type flag uintptr

const flagIndir = 1 << 7
