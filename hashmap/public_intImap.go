package hashmap

import (
	"unsafe" // #nosec

	"github.com/gramework/runtimer"
)

type IntIMap struct {
	hm  *hmap
	typ *runtimer.MapType
}

var intIMapTyp *runtimer.MapType

func init() {
	mi := interface{}(map[int]interface{}{})
	e := *(*emptyInterface)(unsafe.Pointer(&mi))
	intIMapTyp = (*runtimer.MapType)(unsafe.Pointer(e.typ))
}

func NewIntIMap(size ...int32) *IntIMap {
	sz := int32(0)
	if len(size) > 0 {
		sz = size[0]
	}
	typ := &*intIMapTyp
	return &IntIMap{
		typ: typ,
		hm:  makemap(typ, int64(sz), nil, nil),
	}
}

func LoadIntIMap(m interface{}) (*IntIMap, error) {
	if m == nil {
		return nil, ErrNoData
	}
	mi := interface{}(m)
	e := *(*emptyInterface)(unsafe.Pointer(&mi))
	loadedmap := &IntIMap{
		typ: (*runtimer.MapType)(unsafe.Pointer(e.typ)),
		hm:  (*hmap)(e.word),
	}

	return loadedmap, nil
}

func (m *IntIMap) KeyType() string {
	return m.typ.Key.String()
}

func (m *IntIMap) GetPtr(key int) unsafe.Pointer {
	return mapaccess1(m.typ, m.hm, unsafe.Pointer(&key))
}

func (m *IntIMap) GetPtrOk(key int) (unsafe.Pointer, bool) {
	return mapaccess2(m.typ, m.hm, unsafe.Pointer(&key))
}

func (m *IntIMap) Put(key int, value interface{}) {
	p := mapassign(m.typ, m.hm, unsafe.Pointer(&key))
	runtimer.Typedmemmove(m.typ.Elem, p, *runtimer.EfaceDataPtr(&value))
}
