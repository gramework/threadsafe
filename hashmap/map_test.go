package hashmap

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkBigStr_0(b *testing.B)    { benchmarkBigStr(b, 0, false) }
func BenchmarkBigStr_4(b *testing.B)    { benchmarkBigStr(b, 4, false) }
func BenchmarkBigStr_8(b *testing.B)    { benchmarkBigStr(b, 8, false) }
func BenchmarkBigStr_16(b *testing.B)   { benchmarkBigStr(b, 16, false) }
func BenchmarkBigStr_32(b *testing.B)   { benchmarkBigStr(b, 32, false) }
func BenchmarkBigStr_64(b *testing.B)   { benchmarkBigStr(b, 64, false) }
func BenchmarkBigStr_512(b *testing.B)  { benchmarkBigStr(b, 512, false) }
func BenchmarkBigStr2_0(b *testing.B)   { benchmarkBigStr(b, 0, true) }
func BenchmarkBigStr2_4(b *testing.B)   { benchmarkBigStr(b, 4, true) }
func BenchmarkBigStr2_8(b *testing.B)   { benchmarkBigStr(b, 8, true) }
func BenchmarkBigStr2_16(b *testing.B)  { benchmarkBigStr(b, 16, true) }
func BenchmarkBigStr2_32(b *testing.B)  { benchmarkBigStr(b, 32, true) }
func BenchmarkBigStr2_64(b *testing.B)  { benchmarkBigStr(b, 64, true) }
func BenchmarkBigStr2_512(b *testing.B) { benchmarkBigStr(b, 512, true) }

func benchmarkBigStr(b *testing.B, keys int, two bool) {
	m := make(map[string]bool)
	for i := 0; i < keys; i++ {
		suffix := fmt.Sprint(i)
		key := strings.Repeat("X", 1<<20-len(suffix)) + suffix
		m[key] = true
	}
	key := strings.Repeat("X", 1<<20-1) + "k"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if two {
			_, _ = m[key]
		} else {
			_ = m[key]
		}
	}
}

func BenchmarkSmallStr_0(b *testing.B)     { benchmarkSmallStr(b, 0, false) }
func BenchmarkSmallStr_4(b *testing.B)     { benchmarkSmallStr(b, 4, false) }
func BenchmarkSmallStr_8(b *testing.B)     { benchmarkSmallStr(b, 8, false) }
func BenchmarkSmallStr_16(b *testing.B)    { benchmarkSmallStr(b, 16, false) }
func BenchmarkSmallStr_32(b *testing.B)    { benchmarkSmallStr(b, 32, false) }
func BenchmarkSmallStr_64(b *testing.B)    { benchmarkSmallStr(b, 64, false) }
func BenchmarkSmallStr_512(b *testing.B)   { benchmarkSmallStr(b, 512, false) }
func BenchmarkSmallStr_1024(b *testing.B)  { benchmarkSmallStr(b, 1024, false) }
func BenchmarkSmallStr_1M(b *testing.B)    { benchmarkSmallStr(b, 1<<20, false) }
func BenchmarkSmallStr2_0(b *testing.B)    { benchmarkSmallStr(b, 0, true) }
func BenchmarkSmallStr2_4(b *testing.B)    { benchmarkSmallStr(b, 4, true) }
func BenchmarkSmallStr2_8(b *testing.B)    { benchmarkSmallStr(b, 8, true) }
func BenchmarkSmallStr2_16(b *testing.B)   { benchmarkSmallStr(b, 16, true) }
func BenchmarkSmallStr2_32(b *testing.B)   { benchmarkSmallStr(b, 32, true) }
func BenchmarkSmallStr2_64(b *testing.B)   { benchmarkSmallStr(b, 64, true) }
func BenchmarkSmallStr2_512(b *testing.B)  { benchmarkSmallStr(b, 512, true) }
func BenchmarkSmallStr2_1024(b *testing.B) { benchmarkSmallStr(b, 1024, true) }
func BenchmarkSmallStr2_1M(b *testing.B)   { benchmarkSmallStr(b, 1<<20, true) }

func benchmarkSmallStr(b *testing.B, keys int, two bool) {
	m := make(map[string]bool)
	for i := 0; i < keys; i++ {
		m[fmt.Sprint(i)] = true
	}
	b.ResetTimer()
	key := fmt.Sprint(keys + 1)
	for i := 0; i < b.N; i++ {
		if two {
			_, _ = m[key]
		} else {
			_ = m[key]
		}
	}
}

func BenchmarkInt_0(b *testing.B)     { benchmarkInt(b, 0, false) }
func BenchmarkInt_4(b *testing.B)     { benchmarkInt(b, 4, false) }
func BenchmarkInt_8(b *testing.B)     { benchmarkInt(b, 8, false) }
func BenchmarkInt_16(b *testing.B)    { benchmarkInt(b, 16, false) }
func BenchmarkInt_32(b *testing.B)    { benchmarkInt(b, 32, false) }
func BenchmarkInt_64(b *testing.B)    { benchmarkInt(b, 64, false) }
func BenchmarkInt_512(b *testing.B)   { benchmarkInt(b, 512, false) }
func BenchmarkInt_1024(b *testing.B)  { benchmarkInt(b, 1024, false) }
func BenchmarkInt_1M(b *testing.B)    { benchmarkInt(b, 1<<20, false) }
func BenchmarkInt2_0(b *testing.B)    { benchmarkInt(b, 0, true) }
func BenchmarkInt2_4(b *testing.B)    { benchmarkInt(b, 4, true) }
func BenchmarkInt2_8(b *testing.B)    { benchmarkInt(b, 8, true) }
func BenchmarkInt2_16(b *testing.B)   { benchmarkInt(b, 16, true) }
func BenchmarkInt2_32(b *testing.B)   { benchmarkInt(b, 32, true) }
func BenchmarkInt2_64(b *testing.B)   { benchmarkInt(b, 64, true) }
func BenchmarkInt2_512(b *testing.B)  { benchmarkInt(b, 512, true) }
func BenchmarkInt2_1024(b *testing.B) { benchmarkInt(b, 1024, true) }
func BenchmarkInt2_1M(b *testing.B)   { benchmarkInt(b, 1<<20, true) }

func benchmarkInt(b *testing.B, keys int, two bool) {
	m := make(map[int]bool)
	for i := 0; i < keys; i++ {
		m[i] = true
	}
	b.ResetTimer()
	key := keys + 1
	for i := 0; i < b.N; i++ {
		if two {
			_, _ = m[key]
		} else {
			_ = m[key]
		}
	}
}

func BenchmarkPtr_32(b *testing.B)  { benchmarkPtr(b, 32, false) }
func BenchmarkPtr2_32(b *testing.B) { benchmarkPtr(b, 32, true) }

func benchmarkPtr(b *testing.B, keys int, two bool) {
	type foo struct {
		int
		bool
	}
	m := make(map[*foo]bool)
	for i := 0; i < keys; i++ {
		k := &foo{}
		m[k] = true
	}
	key := &foo{}
	for i := 0; i < b.N; i++ {
		if two {
			_, _ = m[key]
		} else {
			_ = m[key]
		}
	}
}
