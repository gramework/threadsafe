package main

import (
	"log"

	"github.com/gramework/threadsafe/hashmap"
	"github.com/kirillDanshin/dlog"
)

func main() {
	om := map[string]string{"1": "1", "2": "2"}
	m, err := hashmap.LoadMap(om)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	v := *(*string)(m.GetPtr("1"))
	dlog.D(v)

	m.Put("3", "3")
	v = *(*string)(m.GetPtr("3"))
	dlog.D(v)
	m.Put("1", "not a 1")
	v = *(*string)(m.GetPtr("1"))
	dlog.D(v)
	vp, ok := m.GetPtrOk("1")
	v = *(*string)(vp)
	dlog.D(v, ok)

	vp, ok = m.GetPtrOk("345345")
	v = *(*string)(vp)
	dlog.D(v, ok)
}
