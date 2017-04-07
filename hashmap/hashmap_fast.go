// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hashmap

import (
	"unsafe" // #nosec

	"github.com/gramework/runtimer"
)

func mapaccess1_fast32(t *runtimer.MapType, h *hmap, key uint32) unsafe.Pointer {
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0])
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map read and map write")
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table. No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(runtimer.Add(h.buckets, (hash&m)*uintptr(t.Bucketsize)))
		if c := h.oldbuckets; c != nil {
			if !h.sameSizeGrow() {
				// There used to be half as many buckets; mask down one more power of two.
				m >>= 1
			}
			oldb := (*bmap)(runtimer.Add(c, (hash&m)*uintptr(t.Bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			k := *((*uint32)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*4)))
			if k != key {
				continue
			}
			x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
			if x == empty {
				continue
			}
			return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*4+i*uintptr(t.Valuesize))
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(&zeroVal[0])
		}
	}
}

func mapaccess2_fast32(t *runtimer.MapType, h *hmap, key uint32) (unsafe.Pointer, bool) {
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0]), false
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map read and map write")
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table. No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(runtimer.Add(h.buckets, (hash&m)*uintptr(t.Bucketsize)))
		if c := h.oldbuckets; c != nil {
			if !h.sameSizeGrow() {
				// There used to be half as many buckets; mask down one more power of two.
				m >>= 1
			}
			oldb := (*bmap)(runtimer.Add(c, (hash&m)*uintptr(t.Bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			k := *((*uint32)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*4)))
			if k != key {
				continue
			}
			x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
			if x == empty {
				continue
			}
			return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*4+i*uintptr(t.Valuesize)), true
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(&zeroVal[0]), false
		}
	}
}

func mapaccess1_fast64(t *runtimer.MapType, h *hmap, key uint64) unsafe.Pointer {
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0])
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map read and map write")
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table. No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(runtimer.Add(h.buckets, (hash&m)*uintptr(t.Bucketsize)))
		if c := h.oldbuckets; c != nil {
			if !h.sameSizeGrow() {
				// There used to be half as many buckets; mask down one more power of two.
				m >>= 1
			}
			oldb := (*bmap)(runtimer.Add(c, (hash&m)*uintptr(t.Bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			k := *((*uint64)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*8)))
			if k != key {
				continue
			}
			x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
			if x == empty {
				continue
			}
			return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.Valuesize))
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(&zeroVal[0])
		}
	}
}

func mapaccess2_fast64(t *runtimer.MapType, h *hmap, key uint64) (unsafe.Pointer, bool) {
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0]), false
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map read and map write")
	}
	var b *bmap
	if h.B == 0 {
		// One-bucket table. No need to hash.
		b = (*bmap)(h.buckets)
	} else {
		hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))
		m := uintptr(1)<<h.B - 1
		b = (*bmap)(runtimer.Add(h.buckets, (hash&m)*uintptr(t.Bucketsize)))
		if c := h.oldbuckets; c != nil {
			if !h.sameSizeGrow() {
				// There used to be half as many buckets; mask down one more power of two.
				m >>= 1
			}
			oldb := (*bmap)(runtimer.Add(c, (hash&m)*uintptr(t.Bucketsize)))
			if !evacuated(oldb) {
				b = oldb
			}
		}
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			k := *((*uint64)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*8)))
			if k != key {
				continue
			}
			x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
			if x == empty {
				continue
			}
			return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.Valuesize)), true
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(&zeroVal[0]), false
		}
	}
}

func mapaccess1_faststr(t *runtimer.MapType, h *hmap, ky string) unsafe.Pointer {
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0])
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map read and map write")
	}
	key := runtimer.StringStructOf(&ky)
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.Len < 32 {
			// short key, doing lots of comparisons is ok
			for i := uintptr(0); i < bucketCnt; i++ {
				x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
				if x == empty {
					continue
				}
				k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*2*runtimer.PtrSize))
				if k.Len != key.Len {
					continue
				}

				if k.Str == key.Str || runtimer.Memequal(k.Str, key.Str, uintptr(key.Len)) {
					return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*runtimer.PtrSize+i*uintptr(t.Valuesize))
				}
			}
			return unsafe.Pointer(&zeroVal[0])
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(bucketCnt)
		for i := uintptr(0); i < bucketCnt; i++ {
			x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
			if x == empty {
				continue
			}
			k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*2*runtimer.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str {
				return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*runtimer.PtrSize+i*uintptr(t.Valuesize))
			}
			// check first 4 bytes
			// TODO: on amd64/386 at least, make this compile to one 4-byte comparison instead of
			// four 1-byte comparisons.
			if *((*[4]byte)(key.Str)) != *((*[4]byte)(k.Str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(runtimer.Add(key.Str, uintptr(key.Len)-4))) != *((*[4]byte)(runtimer.Add(k.Str, uintptr(key.Len)-4))) {
				continue
			}
			if keymaybe != bucketCnt {
				// Two keys are potential matches. Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != bucketCnt {
			k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+keymaybe*2*runtimer.PtrSize))
			if runtimer.Memequal(k.Str, key.Str, uintptr(key.Len)) {
				return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*runtimer.PtrSize+keymaybe*uintptr(t.Valuesize))
			}
		}
		return unsafe.Pointer(&zeroVal[0])
	}
dohash:
	hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(runtimer.Add(h.buckets, (hash&m)*uintptr(t.Bucketsize)))
	if c := h.oldbuckets; c != nil {
		if !h.sameSizeGrow() {
			// There used to be half as many buckets; mask down one more power of two.
			m >>= 1
		}
		oldb := (*bmap)(runtimer.Add(c, (hash&m)*uintptr(t.Bucketsize)))
		if !evacuated(oldb) {
			b = oldb
		}
	}
	top := uint8(hash >> (runtimer.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
			if x != top {
				continue
			}
			k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*2*runtimer.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str || runtimer.Memequal(k.Str, key.Str, uintptr(key.Len)) {
				return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*runtimer.PtrSize+i*uintptr(t.Valuesize))
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(&zeroVal[0])
		}
	}
}

func mapaccess2_faststr(t *runtimer.MapType, h *hmap, ky string) (unsafe.Pointer, bool) {
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0]), false
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map read and map write")
	}
	key := runtimer.StringStructOf(&ky)
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.Len < 32 {
			// short key, doing lots of comparisons is ok
			for i := uintptr(0); i < bucketCnt; i++ {
				x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
				if x == empty {
					continue
				}
				k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*2*runtimer.PtrSize))
				if k.Len != key.Len {
					continue
				}
				if k.Str == key.Str || runtimer.Memequal(k.Str, key.Str, uintptr(key.Len)) {
					return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*runtimer.PtrSize+i*uintptr(t.Valuesize)), true
				}
			}
			return unsafe.Pointer(&zeroVal[0]), false
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(bucketCnt)
		for i := uintptr(0); i < bucketCnt; i++ {
			x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
			if x == empty {
				continue
			}
			k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*2*runtimer.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str {
				return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*runtimer.PtrSize+i*uintptr(t.Valuesize)), true
			}
			// check first 4 bytes
			if *((*[4]byte)(key.Str)) != *((*[4]byte)(k.Str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(runtimer.Add(key.Str, uintptr(key.Len)-4))) != *((*[4]byte)(runtimer.Add(k.Str, uintptr(key.Len)-4))) {
				continue
			}
			if keymaybe != bucketCnt {
				// Two keys are potential matches. Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != bucketCnt {
			k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+keymaybe*2*runtimer.PtrSize))
			if runtimer.Memequal(k.Str, key.Str, uintptr(key.Len)) {
				return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*runtimer.PtrSize+keymaybe*uintptr(t.Valuesize)), true
			}
		}
		return unsafe.Pointer(&zeroVal[0]), false
	}
dohash:
	hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(runtimer.Add(h.buckets, (hash&m)*uintptr(t.Bucketsize)))
	if c := h.oldbuckets; c != nil {
		if !h.sameSizeGrow() {
			// There used to be half as many buckets; mask down one more power of two.
			m >>= 1
		}
		oldb := (*bmap)(runtimer.Add(c, (hash&m)*uintptr(t.Bucketsize)))
		if !evacuated(oldb) {
			b = oldb
		}
	}
	top := uint8(hash >> (runtimer.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			x := *((*uint8)(runtimer.Add(unsafe.Pointer(b), i))) // b.tophash[i] without the bounds check
			if x != top {
				continue
			}
			k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*2*runtimer.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str == key.Str || runtimer.Memequal(k.Str, key.Str, uintptr(key.Len)) {
				return runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*runtimer.PtrSize+i*uintptr(t.Valuesize)), true
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(&zeroVal[0]), false
		}
	}
}

func mapassign_fast32(t *runtimer.MapType, h *hmap, key uint32) unsafe.Pointer {
	if h == nil {
		panic("assignment to entry in nil map")
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map writes")
	}
	hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))

	// Set hashWriting after calling alg.Hash for consistency with mapassign.
	h.flags |= hashWriting

	if h.buckets == nil {
		h.buckets = runtimer.Newarray(runtimer.PtrToType(unsafe.Pointer(t.Bucket)), 1)
	}

again:
	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.growing() {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.Bucketsize)))
	top := uint8(hash >> (runtimer.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}

	var inserti *uint8
	var insertk unsafe.Pointer
	var val unsafe.Pointer
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			if b.tophash[i] != top {
				if b.tophash[i] == empty && inserti == nil {
					inserti = &b.tophash[i]
					insertk = runtimer.Add(unsafe.Pointer(b), dataOffset+i*4)
					val = runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*4+i*uintptr(t.Valuesize))
				}
				continue
			}
			k := *((*uint32)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*4)))
			if k != key {
				continue
			}
			val = runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*4+i*uintptr(t.Valuesize))
			goto done
		}
		ovf := b.overflow(t)
		if ovf == nil {
			break
		}
		b = ovf
	}

	// Did not find mapping for key. Allocate new cell & add entry.

	// If we hit the max load factor or we have too many overflow buckets,
	// and we're not already in the middle of growing, start growing.
	if !h.growing() && (overLoadFactor(int64(h.count), h.B) || tooManyOverflowBuckets(h.noverflow, h.B)) {
		hashGrow(t, h)
		goto again // Growing the table invalidates everything, so try again
	}

	if inserti == nil {
		// all current buckets are full, allocate a new one.
		newb := (*bmap)(runtimer.Newobject(runtimer.PtrToType(unsafe.Pointer(t.Bucket))))
		h.setoverflow(t, b, newb)
		inserti = &newb.tophash[0]
		insertk = runtimer.Add(unsafe.Pointer(newb), dataOffset)
		val = runtimer.Add(insertk, bucketCnt*4)
	}

	// store new key/value at insert position
	*((*uint32)(insertk)) = key
	*inserti = top
	h.count++

done:
	if h.flags&hashWriting == 0 {
		runtimer.Throw("concurrent map writes")
	}
	h.flags &^= hashWriting
	return val
}

func mapassign_fast64(t *runtimer.MapType, h *hmap, key uint64) unsafe.Pointer {
	if h == nil {
		panic("assignment to entry in nil map")
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map writes")
	}
	hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))

	// Set hashWriting after calling alg.Hash for consistency with mapassign.
	h.flags |= hashWriting

	if h.buckets == nil {
		h.buckets = runtimer.Newarray(runtimer.PtrToType(unsafe.Pointer(t.Bucket)), 1)
	}

again:
	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.growing() {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.Bucketsize)))
	top := uint8(hash >> (runtimer.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}

	var inserti *uint8
	var insertk unsafe.Pointer
	var val unsafe.Pointer
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			if b.tophash[i] != top {
				if b.tophash[i] == empty && inserti == nil {
					inserti = &b.tophash[i]
					insertk = runtimer.Add(unsafe.Pointer(b), dataOffset+i*8)
					val = runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.Valuesize))
				}
				continue
			}
			k := *((*uint64)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*8)))
			if k != key {
				continue
			}
			val = runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*8+i*uintptr(t.Valuesize))
			goto done
		}
		ovf := b.overflow(t)
		if ovf == nil {
			break
		}
		b = ovf
	}

	// Did not find mapping for key. Allocate new cell & add entry.

	// If we hit the max load factor or we have too many overflow buckets,
	// and we're not already in the middle of growing, start growing.
	if !h.growing() && (overLoadFactor(int64(h.count), h.B) || tooManyOverflowBuckets(h.noverflow, h.B)) {
		hashGrow(t, h)
		goto again // Growing the table invalidates everything, so try again
	}

	if inserti == nil {
		// all current buckets are full, allocate a new one.
		newb := (*bmap)(runtimer.Newobject(runtimer.PtrToType(unsafe.Pointer(t.Bucket))))
		h.setoverflow(t, b, newb)
		inserti = &newb.tophash[0]
		insertk = runtimer.Add(unsafe.Pointer(newb), dataOffset)
		val = runtimer.Add(insertk, bucketCnt*8)
	}

	// store new key/value at insert position
	*((*uint64)(insertk)) = key
	*inserti = top
	h.count++

done:
	if h.flags&hashWriting == 0 {
		runtimer.Throw("concurrent map writes")
	}
	h.flags &^= hashWriting
	return val
}

func mapassign_faststr(t *runtimer.MapType, h *hmap, ky string) unsafe.Pointer {
	if h == nil {
		panic("assignment to entry in nil map")
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map writes")
	}
	key := runtimer.StringStructOf(&ky)
	hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))

	// Set hashWriting after calling alg.Hash for consistency with mapassign.
	h.flags |= hashWriting

	if h.buckets == nil {
		h.buckets = runtimer.Newarray(runtimer.PtrToType(unsafe.Pointer(t.Bucket)), 1)
	}

again:
	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.growing() {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.Bucketsize)))
	top := uint8(hash >> (runtimer.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}

	var inserti *uint8
	var insertk unsafe.Pointer
	var val unsafe.Pointer
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			if b.tophash[i] != top {
				if b.tophash[i] == empty && inserti == nil {
					inserti = &b.tophash[i]
					insertk = runtimer.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.Keysize))
					val = runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.Keysize)+i*uintptr(t.Valuesize))
				}
				continue
			}
			k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*2*runtimer.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str != key.Str && !runtimer.Memequal(k.Str, key.Str, uintptr(key.Len)) {
				continue
			}
			// already have a mapping for key. Update it.
			val = runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*2*runtimer.PtrSize+i*uintptr(t.Valuesize))
			goto done
		}
		ovf := b.overflow(t)
		if ovf == nil {
			break
		}
		b = ovf
	}

	// Did not find mapping for key. Allocate new cell & add entry.

	// If we hit the max load factor or we have too many overflow buckets,
	// and we're not already in the middle of growing, start growing.
	if !h.growing() && (overLoadFactor(int64(h.count), h.B) || tooManyOverflowBuckets(h.noverflow, h.B)) {
		hashGrow(t, h)
		goto again // Growing the table invalidates everything, so try again
	}

	if inserti == nil {
		// all current buckets are full, allocate a new one.
		newb := (*bmap)(runtimer.Newobject(runtimer.PtrToType(unsafe.Pointer(t.Bucket))))
		h.setoverflow(t, b, newb)
		inserti = &newb.tophash[0]
		insertk = runtimer.Add(unsafe.Pointer(newb), dataOffset)
		val = runtimer.Add(insertk, bucketCnt*2*runtimer.PtrSize)
	}

	// store new key/value at insert position
	*((*runtimer.StringStruct)(insertk)) = *key
	*inserti = top
	h.count++

done:
	if h.flags&hashWriting == 0 {
		runtimer.Throw("concurrent map writes")
	}
	h.flags &^= hashWriting
	return val
}

func mapdelete_fast32(t *runtimer.MapType, h *hmap, key uint32) {
	if h == nil || h.count == 0 {
		return
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map writes")
	}

	hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))

	// Set hashWriting after calling alg.Hash for consistency with mapdelete
	h.flags |= hashWriting

	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.growing() {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.Bucketsize)))
	top := uint8(hash >> (runtimer.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			if b.tophash[i] != top {
				continue
			}
			k := (*uint32)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*4))
			if key != *k {
				continue
			}
			*k = 0
			v := unsafe.Pointer(uintptr(unsafe.Pointer(b)) + dataOffset + bucketCnt*4 + i*uintptr(t.Valuesize))
			runtimer.Typedmemclr(runtimer.PtrToType(unsafe.Pointer(t.Elem)), v)
			b.tophash[i] = empty
			h.count--
			goto done
		}
		b = b.overflow(t)
		if b == nil {
			goto done
		}
	}

done:
	if h.flags&hashWriting == 0 {
		runtimer.Throw("concurrent map writes")
	}
	h.flags &^= hashWriting
}

func mapdelete_fast64(t *runtimer.MapType, h *hmap, key uint64) {
	if h == nil || h.count == 0 {
		return
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map writes")
	}

	hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&key)), uintptr(h.hash0))

	// Set hashWriting after calling alg.Hash for consistency with mapdelete
	h.flags |= hashWriting

	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.growing() {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.Bucketsize)))
	top := uint8(hash >> (runtimer.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			if b.tophash[i] != top {
				continue
			}
			k := (*uint64)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*8))
			if key != *k {
				continue
			}
			*k = 0
			v := unsafe.Pointer(uintptr(unsafe.Pointer(b)) + dataOffset + bucketCnt*8 + i*uintptr(t.Valuesize))
			runtimer.Typedmemclr(runtimer.PtrToType(unsafe.Pointer(t.Elem)), v)
			b.tophash[i] = empty
			h.count--
			goto done
		}
		b = b.overflow(t)
		if b == nil {
			goto done
		}
	}

done:
	if h.flags&hashWriting == 0 {
		runtimer.Throw("concurrent map writes")
	}
	h.flags &^= hashWriting
}

func mapdelete_faststr(t *runtimer.MapType, h *hmap, ky string) {
	if h == nil || h.count == 0 {
		return
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map writes")
	}

	key := runtimer.StringStructOf(&ky)
	hash := t.Key.Alg.Hash(runtimer.Noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))

	// Set hashWriting after calling alg.Hash for consistency with mapdelete
	h.flags |= hashWriting

	bucket := hash & (uintptr(1)<<h.B - 1)
	if h.growing() {
		growWork(t, h, bucket)
	}
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.Bucketsize)))
	top := uint8(hash >> (runtimer.PtrSize*8 - 8))
	if top < minTopHash {
		top += minTopHash
	}
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			if b.tophash[i] != top {
				continue
			}
			k := (*runtimer.StringStruct)(runtimer.Add(unsafe.Pointer(b), dataOffset+i*2*runtimer.PtrSize))
			if k.Len != key.Len {
				continue
			}
			if k.Str != key.Str && !runtimer.Memequal(k.Str, key.Str, uintptr(key.Len)) {
				continue
			}
			runtimer.Typedmemclr(runtimer.PtrToType(unsafe.Pointer(t.Key)), unsafe.Pointer(k))
			v := unsafe.Pointer(uintptr(unsafe.Pointer(b)) + dataOffset + bucketCnt*2*runtimer.PtrSize + i*uintptr(t.Valuesize))
			runtimer.Typedmemclr(runtimer.PtrToType(unsafe.Pointer(t.Elem)), v)
			b.tophash[i] = empty
			h.count--
			goto done
		}
		b = b.overflow(t)
		if b == nil {
			goto done
		}
	}

done:
	if h.flags&hashWriting == 0 {
		runtimer.Throw("concurrent map writes")
	}
	h.flags &^= hashWriting
}
