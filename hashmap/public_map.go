package hashmap

import (
	"errors"
	"unsafe" // #nosec

	"github.com/gramework/runtimer"
)

var ErrNoType = errors.New("no type can be loaded")
var ErrNoData = errors.New("nil map, no data can be loaded")
var ErrNotAMap = errors.New("map should be passed by value to LoadMap()")

type Map struct {
	hm  *hmap
	typ *runtimer.MapType
}

func LoadMap(m interface{}) (*Map, error) {
	e := *(*emptyInterface)(unsafe.Pointer(&m))
	if e.typ == nil {
		return nil, ErrNoType
	}
	if e.word == nil {
		return nil, ErrNoData
	}

	if (*runtimer.MapType)(unsafe.Pointer(e.typ)).Key.Alg.Hash == nil {
		return nil, ErrNotAMap
	}
	loadedmap := &Map{
		typ: (*runtimer.MapType)(unsafe.Pointer(e.typ)),
		hm:  (*hmap)(e.word),
	}

	return loadedmap, nil
}

func (m *Map) KeyType() string {
	return m.typ.Key.String()
}

func (m *Map) GetPtr(key interface{}) unsafe.Pointer {
	return mapaccess1(m.typ, m.hm, runtimer.GetEfaceDataPtr(key))
}

func (m *Map) GetPtrOk(key interface{}) (unsafe.Pointer, bool) {
	return mapaccess2(m.typ, m.hm, runtimer.GetEfaceDataPtr(key))
}

func (m *Map) Put(key, value interface{}) {
	p := mapassign(m.typ, m.hm, runtimer.GetEfaceDataPtr(key))
	runtimer.Typedmemmove(m.typ.Elem, p, runtimer.GetEfaceDataPtr(value))
}
