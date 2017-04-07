package hashmap

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkHashMapBigStr_0(b *testing.B)    { benchmarkHashMapBigStr(b, 0, false) }
func BenchmarkHashMapBigStr_4(b *testing.B)    { benchmarkHashMapBigStr(b, 4, false) }
func BenchmarkHashMapBigStr_8(b *testing.B)    { benchmarkHashMapBigStr(b, 8, false) }
func BenchmarkHashMapBigStr_16(b *testing.B)   { benchmarkHashMapBigStr(b, 16, false) }
func BenchmarkHashMapBigStr_32(b *testing.B)   { benchmarkHashMapBigStr(b, 32, false) }
func BenchmarkHashMapBigStr_64(b *testing.B)   { benchmarkHashMapBigStr(b, 64, false) }
func BenchmarkHashMapBigStr_512(b *testing.B)  { benchmarkHashMapBigStr(b, 512, false) }
func BenchmarkHashMapBigStr2_0(b *testing.B)   { benchmarkHashMapBigStr(b, 0, true) }
func BenchmarkHashMapBigStr2_4(b *testing.B)   { benchmarkHashMapBigStr(b, 4, true) }
func BenchmarkHashMapBigStr2_8(b *testing.B)   { benchmarkHashMapBigStr(b, 8, true) }
func BenchmarkHashMapBigStr2_16(b *testing.B)  { benchmarkHashMapBigStr(b, 16, true) }
func BenchmarkHashMapBigStr2_32(b *testing.B)  { benchmarkHashMapBigStr(b, 32, true) }
func BenchmarkHashMapBigStr2_64(b *testing.B)  { benchmarkHashMapBigStr(b, 64, true) }
func BenchmarkHashMapBigStr2_512(b *testing.B) { benchmarkHashMapBigStr(b, 512, true) }

func benchmarkHashMapBigStr(b *testing.B, keys int, two bool) {
	m := NewStrIMap()

	for i := 0; i < keys; i++ {
		suffix := fmt.Sprint(i)
		key := strings.Repeat("X", 1<<20-len(suffix)) + suffix
		m.Put(key, true)
	}
	key := strings.Repeat("X", 1<<20-1) + "k"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if two {
			_, _ = m.GetPtrOk(key)
		} else {
			_ = m.GetPtr(key)
		}
	}
}

func BenchmarkHashMapSmallStr_0(b *testing.B)     { benchmarkHashMapSmallStr(b, 0, false) }
func BenchmarkHashMapSmallStr_4(b *testing.B)     { benchmarkHashMapSmallStr(b, 4, false) }
func BenchmarkHashMapSmallStr_8(b *testing.B)     { benchmarkHashMapSmallStr(b, 8, false) }
func BenchmarkHashMapSmallStr_16(b *testing.B)    { benchmarkHashMapSmallStr(b, 16, false) }
func BenchmarkHashMapSmallStr_32(b *testing.B)    { benchmarkHashMapSmallStr(b, 32, false) }
func BenchmarkHashMapSmallStr_64(b *testing.B)    { benchmarkHashMapSmallStr(b, 64, false) }
func BenchmarkHashMapSmallStr_512(b *testing.B)   { benchmarkHashMapSmallStr(b, 512, false) }
func BenchmarkHashMapSmallStr_1024(b *testing.B)  { benchmarkHashMapSmallStr(b, 1024, false) }
func BenchmarkHashMapSmallStr_1M(b *testing.B)    { benchmarkHashMapSmallStr(b, 1<<20, false) }
func BenchmarkHashMapSmallStr2_0(b *testing.B)    { benchmarkHashMapSmallStr(b, 0, true) }
func BenchmarkHashMapSmallStr2_4(b *testing.B)    { benchmarkHashMapSmallStr(b, 4, true) }
func BenchmarkHashMapSmallStr2_8(b *testing.B)    { benchmarkHashMapSmallStr(b, 8, true) }
func BenchmarkHashMapSmallStr2_16(b *testing.B)   { benchmarkHashMapSmallStr(b, 16, true) }
func BenchmarkHashMapSmallStr2_32(b *testing.B)   { benchmarkHashMapSmallStr(b, 32, true) }
func BenchmarkHashMapSmallStr2_64(b *testing.B)   { benchmarkHashMapSmallStr(b, 64, true) }
func BenchmarkHashMapSmallStr2_512(b *testing.B)  { benchmarkHashMapSmallStr(b, 512, true) }
func BenchmarkHashMapSmallStr2_1024(b *testing.B) { benchmarkHashMapSmallStr(b, 1024, true) }
func BenchmarkHashMapSmallStr2_1M(b *testing.B)   { benchmarkHashMapSmallStr(b, 1<<20, true) }

func benchmarkHashMapSmallStr(b *testing.B, keys int, two bool) {
	m := NewStrIMap()
	for i := 0; i < keys; i++ {
		m.Put(fmt.Sprint(i), true)
	}
	key := fmt.Sprint(keys + 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if two {
			_, _ = m.GetPtrOk(key)
		} else {
			_ = m.GetPtr(key)
		}
	}
}

func BenchmarkHashMapInt_0(b *testing.B)     { benchmarkHashMapInt(b, 0, false) }
func BenchmarkHashMapInt_4(b *testing.B)     { benchmarkHashMapInt(b, 4, false) }
func BenchmarkHashMapInt_8(b *testing.B)     { benchmarkHashMapInt(b, 8, false) }
func BenchmarkHashMapInt_16(b *testing.B)    { benchmarkHashMapInt(b, 16, false) }
func BenchmarkHashMapInt_32(b *testing.B)    { benchmarkHashMapInt(b, 32, false) }
func BenchmarkHashMapInt_64(b *testing.B)    { benchmarkHashMapInt(b, 64, false) }
func BenchmarkHashMapInt_512(b *testing.B)   { benchmarkHashMapInt(b, 512, false) }
func BenchmarkHashMapInt_1024(b *testing.B)  { benchmarkHashMapInt(b, 1024, false) }
func BenchmarkHashMapInt_1M(b *testing.B)    { benchmarkHashMapInt(b, 1<<20, false) }
func BenchmarkHashMapInt2_0(b *testing.B)    { benchmarkHashMapInt(b, 0, true) }
func BenchmarkHashMapInt2_4(b *testing.B)    { benchmarkHashMapInt(b, 4, true) }
func BenchmarkHashMapInt2_8(b *testing.B)    { benchmarkHashMapInt(b, 8, true) }
func BenchmarkHashMapInt2_16(b *testing.B)   { benchmarkHashMapInt(b, 16, true) }
func BenchmarkHashMapInt2_32(b *testing.B)   { benchmarkHashMapInt(b, 32, true) }
func BenchmarkHashMapInt2_64(b *testing.B)   { benchmarkHashMapInt(b, 64, true) }
func BenchmarkHashMapInt2_512(b *testing.B)  { benchmarkHashMapInt(b, 512, true) }
func BenchmarkHashMapInt2_1024(b *testing.B) { benchmarkHashMapInt(b, 1024, true) }
func BenchmarkHashMapInt2_1M(b *testing.B)   { benchmarkHashMapInt(b, 1<<20, true) }

func benchmarkHashMapInt(b *testing.B, keys int, two bool) {
	m := NewIntIMap()
	for i := 0; i < keys; i++ {
		m.Put(i, true)
	}
	b.ResetTimer()
	key := keys + 1
	for i := 0; i < b.N; i++ {
		if two {
			_, _ = m.GetPtrOk(key)
		} else {
			_ = m.GetPtr(key)
		}
	}
}

func BenchmarkHashMapPtr_32(b *testing.B)  { benchmarkHashMapPtr(b, 32, false) }
func BenchmarkHashMapPtr2_32(b *testing.B) { benchmarkHashMapPtr(b, 32, true) }

func benchmarkHashMapPtr(b *testing.B, keys int, two bool) {
	type foo struct {
		int
		bool
	}
	m, err := LoadMap(map[*foo]bool{})
	if err != nil {
		b.Errorf("Can't load map: %s", err)
		b.FailNow()
	}
	for i := 0; i < keys; i++ {
		k := &foo{}
		m.Put(k, true)
	}
	key := &foo{}
	for i := 0; i < b.N; i++ {
		if two {
			_, _ = m.GetPtrOk(key)
		} else {
			_ = m.GetPtr(key)
		}
	}
}
