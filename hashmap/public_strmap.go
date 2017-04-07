package hashmap

import (
	"unsafe" // #nosec

	"github.com/gramework/runtimer"
)

type StrMap struct {
	hm  *hmap
	typ *runtimer.MapType
}

var strMapTyp *runtimer.MapType

func init() {
	mi := interface{}(map[string]string{})
	e := *(*emptyInterface)(unsafe.Pointer(&mi))
	strMapTyp = (*runtimer.MapType)(unsafe.Pointer(e.typ))
}

func NewStrMap() *StrMap {
	typ := &*strMapTyp
	return &StrMap{
		typ: typ,
		hm:  makemap(typ, 0, nil, nil),
	}
}

func LoadStrMap(m map[string]string) (*StrMap, error) {
	if m == nil {
		return nil, ErrNoData
	}
	mi := interface{}(m)
	e := *(*emptyInterface)(unsafe.Pointer(&mi))
	loadedmap := &StrMap{
		typ: (*runtimer.MapType)(unsafe.Pointer(e.typ)),
		hm:  (*hmap)(e.word),
	}

	return loadedmap, nil
}

func (m *StrMap) KeyType() string {
	return m.typ.Key.String()
}

func (m *StrMap) GetPtr(key string) unsafe.Pointer {
	return mapaccess1_faststr(m.typ, m.hm, *runtimer.PtrToStringPtr(runtimer.GetEfaceDataPtr(&key)))
}

func (m *StrMap) GetPtrOk(key string) (unsafe.Pointer, bool) {
	return mapaccess2_faststr(m.typ, m.hm, key)
}

func (m *StrMap) Put(key, value string) {
	p := mapassign_faststr(m.typ, m.hm, key)
	runtimer.Typedmemmove(m.typ.Elem, p, unsafe.Pointer(&value))
}
