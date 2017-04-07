package hashmap

import (
	"unsafe" // #nosec

	"github.com/gramework/runtimer"
)

type StrIMap struct {
	hm  *hmap
	typ *runtimer.MapType
}

var strIMapTyp *runtimer.MapType

func init() {
	mi := interface{}(map[string]interface{}{})
	e := *(*emptyInterface)(unsafe.Pointer(&mi))
	strIMapTyp = (*runtimer.MapType)(unsafe.Pointer(e.typ))
}

func NewStrIMap() *StrIMap {
	typ := &*strIMapTyp
	return &StrIMap{
		typ: typ,
		hm:  makemap(typ, 0, nil, nil),
	}
}

func LoadStrIMap(m interface{}) (*StrIMap, error) {
	if m == nil {
		return nil, ErrNoData
	}
	mi := interface{}(m)
	e := *(*emptyInterface)(unsafe.Pointer(&mi))
	loadedmap := &StrIMap{
		typ: (*runtimer.MapType)(unsafe.Pointer(e.typ)),
		hm:  (*hmap)(e.word),
	}

	return loadedmap, nil
}

func (m *StrIMap) KeyType() string {
	return m.typ.Key.String()
}

func (m *StrIMap) GetPtr(key string) unsafe.Pointer {
	return mapaccess1_faststr(m.typ, m.hm, key)
}

func (m *StrIMap) GetPtrOk(key string) (unsafe.Pointer, bool) {
	return mapaccess2_faststr(m.typ, m.hm, key)
}

func (m *StrIMap) Put(key string, value interface{}) {
	p := mapassign(m.typ, m.hm, unsafe.Pointer(&key))
	runtimer.Typedmemmove(m.typ.Elem, p, *runtimer.EfaceDataPtr(&value))
}
