// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hashmap

// This file contains the implementation of Go's map type.
//
// A map is just a hash table. The data is arranged
// into an array of buckets. Each bucket contains up to
// 8 key/value pairs. The low-order bits of the hash are
// used to select a bucket. Each bucket contains a few
// high-order bits of each hash to distinguish the entries
// within a single bucket.
//
// If more than 8 keys hash to a bucket, we chain on
// extra buckets.
//
// When the hashtable grows, we allocate a new array
// of buckets twice as big. Buckets are incrementally
// copied from the old bucket array to the new bucket array.
//
// Map iterators walk through the array of buckets and
// return the keys in walk order (bucket #, then overflow
// chain order, then bucket index).  To maintain iteration
// semantics, we never move keys within their bucket (if
// we did, keys might be returned 0 or 2 times).  When
// growing the table, iterators remain iterating through the
// old table and must check the new table if the bucket
// they are iterating through has been moved ("evacuated")
// to the new table.

// Picking loadFactor: too large and we have lots of overflow
// buckets, too small and we waste a lot of space. I wrote
// a simple program to check some stats for different loads:
// (64-bit, 8 byte keys and values)
//  loadFactor    %overflow  bytes/entry     hitprobe    missprobe
//        4.00         2.13        20.77         3.00         4.00
//        4.50         4.05        17.30         3.25         4.50
//        5.00         6.85        14.77         3.50         5.00
//        5.50        10.55        12.94         3.75         5.50
//        6.00        15.27        11.67         4.00         6.00
//        6.50        20.90        10.79         4.25         6.50
//        7.00        27.14        10.15         4.50         7.00
//        7.50        34.03         9.73         4.75         7.50
//        8.00        41.10         9.40         5.00         8.00
//
// %overflow   = percentage of buckets which have an overflow bucket
// bytes/entry = overhead bytes used per key/value pair
// hitprobe    = # of entries to check when looking up a present key
// missprobe   = # of entries to check when looking up an absent key
//
// Keep in mind this data is for maximally loaded tables, i.e. just
// before the table grows. Typical tables will be somewhat less loaded.

import (
	"unsafe" // #nosec

	"github.com/gramework/runtimer"
)

const (
	// Maximum number of key/value pairs a bucket can hold.
	bucketCntBits = 3
	bucketCnt     = 1 << bucketCntBits

	// Maximum average load of a bucket that triggers growth.
	loadFactor = 6.5

	// Maximum key or value size to keep inline (instead of mallocing per element).
	// Must fit in a uint8.
	// Fast versions cannot handle big values - the cutoff size for
	// fast versions in ../../cmd/internal/gc/walk.go must be at most this value.
	maxKeySize   = 128
	maxValueSize = 128

	// data offset should be the size of the bmap struct, but needs to be
	// aligned correctly. For amd64p32 this means 64-bit alignment
	// even though pointers are 32 bit.
	dataOffset = unsafe.Offsetof(struct {
		b bmap
		v int64
	}{}.v)

	// Possible tophash values. We reserve a few possibilities for special marks.
	// Each bucket (including its overflow buckets, if any) will have either all or none of its
	// entries in the evacuated* states (except during the evacuate() method, which only happens
	// during map writes and thus no one else can observe the map during that time).
	empty          = 0 // cell is empty
	evacuatedEmpty = 1 // cell is empty, bucket is evacuated.
	evacuatedX     = 2 // key/value is valid.  Entry has been evacuated to first half of larger table.
	evacuatedY     = 3 // same as above, but evacuated to second half of larger table.
	minTopHash     = 4 // minimum tophash for a normal filled cell.

	// flags
	iterator     = 1 // there may be an iterator using buckets
	oldIterator  = 2 // there may be an iterator using oldbuckets
	hashWriting  = 4 // a goroutine is writing to the map
	sameSizeGrow = 8 // the current map growth is to a new map of the same size

	// sentinel bucket ID for iterator checks
	noCheck = 1<<(8*runtimer.PtrSize) - 1
)

// A header for a Go map.
type hmap struct {
	// Note: the format of the Hmap is encoded in ../../cmd/internal/gc/reflect.go and
	// ../reflect/type.go. Don't change this structure without also changing that code!
	count     int // # live cells == size of map.  Must be first (used by len() builtin)
	flags     uint8
	B         uint8  // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
	noverflow uint16 // approximate number of overflow buckets; see incrnoverflow for details
	hash0     uint32 // hash seed

	buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
	oldbuckets unsafe.Pointer // previous bucket array of half the size, non-nil only when growing
	nevacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated)

	// If both key and value do not contain pointers and are inline, then we mark bucket
	// type as containing no pointers. This avoids scanning such maps.
	// However, bmap.overflow is a pointer. In order to keep overflow buckets
	// alive, we store pointers to all overflow buckets in hmap.overflow.
	// Overflow is used only if key and value do not contain pointers.
	// overflow[0] contains overflow buckets for hmap.buckets.
	// overflow[1] contains overflow buckets for hmap.oldbuckets.
	// The first indirection allows us to reduce static size of hmap.
	// The second indirection allows to store a pointer to the slice in hiter.
	overflow *[2]*[]*bmap
}

// A bucket for a Go map.
type bmap struct {
	// tophash generally contains the top byte of the hash value
	// for each key in this bucket. If tophash[0] < minTopHash,
	// tophash[0] is a bucket evacuation state instead.
	tophash [bucketCnt]uint8
	// Followed by bucketCnt keys and then bucketCnt values.
	// NOTE: packing all the keys together and then all the values together makes the
	// code a bit more complicated than alternating key/value/key/value/... but it allows
	// us to eliminate padding which would be needed for, e.g., map[int64]int8.
	// Followed by an overflow pointer.
}

// A hash iteration structure.
// If you modify hiter, also change cmd/internal/gc/reflect.go to indicate
// the layout of this structure.
type hiter struct {
	key         unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/internal/gc/range.go).
	value       unsafe.Pointer // Must be in second position (see cmd/internal/gc/range.go).
	t           *runtimer.MapType
	h           *hmap
	buckets     unsafe.Pointer // bucket ptr at hash_iter initialization time
	bptr        *bmap          // current bucket
	overflow    [2]*[]*bmap    // keeps overflow buckets alive
	startBucket uintptr        // bucket iteration started at
	offset      uint8          // intra-bucket offset to start from during iteration (should be big enough to hold bucketCnt-1)
	wrapped     bool           // already wrapped around from end of bucket array to beginning
	B           uint8
	i           uint8
	bucket      uintptr
	checkBucket uintptr
}

func evacuated(b *bmap) bool {
	h := b.tophash[0]
	return h > empty && h < minTopHash
}

func (b *bmap) overflow(t *runtimer.MapType) *bmap {
	return *(**bmap)(runtimer.Add(unsafe.Pointer(b), uintptr(t.Bucketsize)-runtimer.PtrSize))
}

// incrnoverflow increments h.noverflow.
// noverflow counts the number of overflow buckets.
// This is used to trigger same-size map growth.
// See also tooManyOverflowBuckets.
// To keep hmap small, noverflow is a uint16.
// When there are few buckets, noverflow is an exact count.
// When there are many buckets, noverflow is an approximate count.
func (h *hmap) incrnoverflow() {
	// We trigger same-size map growth if there are
	// as many overflow buckets as buckets.
	// We need to be able to count to 1<<h.B.
	if h.B < 16 {
		h.noverflow++
		return
	}
	// Increment with probability 1/(1<<(h.B-15)).
	// When we reach 1<<15 - 1, we will have approximately
	// as many overflow buckets as buckets.
	mask := uint32(1)<<(h.B-15) - 1
	// Example: if h.B == 18, then mask == 7,
	// and fastrand & 7 == 0 with probability 1/8.
	if runtimer.Fastrand()&mask == 0 {
		h.noverflow++
	}
}

func (h *hmap) setoverflow(t *runtimer.MapType, b, ovf *bmap) {
	h.incrnoverflow()
	if t.Bucket.Kind&runtimer.KindNoPointers != 0 {
		h.createOverflow()
		*h.overflow[0] = append(*h.overflow[0], ovf)
	}
	*(**bmap)(runtimer.Add(unsafe.Pointer(b), uintptr(t.Bucketsize)-runtimer.PtrSize)) = ovf
}

func (h *hmap) createOverflow() {
	if h.overflow == nil {
		h.overflow = new([2]*[]*bmap)
	}
	if h.overflow[0] == nil {
		h.overflow[0] = new([]*bmap)
	}
}

// makemap implements a Go map creation make(map[k]v, hint)
// If the compiler has determined that the map or the first bucket
// can be created on the stack, h and/or bucket may be non-nil.
// If h != nil, the map can be created directly in h.
// If bucket != nil, bucket can be used as the first bucket.
func makemap(t *runtimer.MapType, hint int64, h *hmap, bucket unsafe.Pointer) *hmap {
	if sz := unsafe.Sizeof(hmap{}); sz > 48 || sz != t.Hmap.Size {
		println("runtimer: sizeof(hmap) =", sz, ", t.Hmap.size =", t.Hmap.Size)
		runtimer.Throw("bad hmap size")
	}

	if hint < 0 || int64(int32(hint)) != hint {
		panic("makemap: size out of range")
		// TODO: make hint an int, then none of this nonsense
	}

	if !ismapkey(t.Key) {
		runtimer.Throw("runtimer.makemap: unsupported map key type")
	}

	// check compiler's and reflect's math
	if t.Key.Size > maxKeySize && (!t.Indirectkey || t.Keysize != uint8(runtimer.PtrSize)) ||
		t.Key.Size <= maxKeySize && (t.Indirectkey || t.Keysize != uint8(t.Key.Size)) {
		runtimer.Throw("key size wrong")
	}
	if t.Elem.Size > maxValueSize && (!t.Indirectvalue || t.Valuesize != uint8(runtimer.PtrSize)) ||
		t.Elem.Size <= maxValueSize && (t.Indirectvalue || t.Valuesize != uint8(t.Elem.Size)) {
		runtimer.Throw("value size wrong")
	}

	// invariants we depend on. We should probably check these at compile time
	// somewhere, but for now we'll do it here.
	if t.Key.Align > bucketCnt {
		runtimer.Throw("key align too big")
	}
	if t.Elem.Align > bucketCnt {
		runtimer.Throw("value align too big")
	}
	if t.Key.Size%uintptr(t.Key.Align) != 0 {
		runtimer.Throw("key size not a multiple of key align")
	}
	if t.Elem.Size%uintptr(t.Elem.Align) != 0 {
		runtimer.Throw("value size not a multiple of value align")
	}
	if bucketCnt < 8 {
		runtimer.Throw("bucketsize too small for proper alignment")
	}
	if dataOffset%uintptr(t.Key.Align) != 0 {
		runtimer.Throw("need padding in bucket (key)")
	}
	if dataOffset%uintptr(t.Elem.Align) != 0 {
		runtimer.Throw("need padding in bucket (value)")
	}

	// find size parameter which will hold the requested # of elements
	B := uint8(0)
	for ; overLoadFactor(hint, B); B++ {
	}

	// allocate initial hash table
	// if B == 0, the buckets field is allocated lazily later (in mapassign)
	// If hint is large zeroing this memory could take a while.
	buckets := bucket
	if B != 0 {
		buckets = runtimer.Newarray((*runtimer.Type)((unsafe.Pointer)(t.Bucket)), 1<<B)
	}

	// initialize Hmap
	if h == nil {
		h = (*hmap)(runtimer.Newobject((*runtimer.Type)((unsafe.Pointer)(t.Hmap))))
	}
	h.count = 0
	h.B = B
	h.flags = 0
	h.hash0 = runtimer.Fastrand()
	h.buckets = buckets
	h.oldbuckets = nil
	h.nevacuate = 0
	h.noverflow = 0

	return h
}

// mapaccess1 returns a pointer to h[key].  Never returns nil, instead
// it will return a reference to the zero object for the value type if
// the key is not in the map.
// NOTE: The returned pointer may keep the whole map live, so don't
// hold onto it for very long.
func mapaccess1(t *runtimer.MapType, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	if runtimer.Msanenabled && h != nil {
		runtimer.Msanread(key, t.Key.Size)
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0])
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map read and map write")
	}
	alg := t.Key.Alg
	if alg.Hash == nil {
		panic("no alg.Hash!")
	}
	hash := alg.Hash(key, uintptr(h.hash0))
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
			if b.tophash[i] != top {
				continue
			}
			k := runtimer.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.Keysize))
			if t.Indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if alg.Equal(key, k) {
				v := runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.Keysize)+i*uintptr(t.Valuesize))
				if t.Indirectvalue {
					v = *((*unsafe.Pointer)(v))
				}
				return v
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(&zeroVal[0])
		}
	}
}

func mapaccess2(t *runtimer.MapType, h *hmap, key unsafe.Pointer) (unsafe.Pointer, bool) {
	if runtimer.Msanenabled && h != nil {
		runtimer.Msanread(key, t.Key.Size)
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0]), false
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map read and map write")
	}
	alg := t.Key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + (hash&m)*uintptr(t.Bucketsize)))
	if c := h.oldbuckets; c != nil {
		if !h.sameSizeGrow() {
			// There used to be half as many buckets; mask down one more power of two.
			m >>= 1
		}
		oldb := (*bmap)(unsafe.Pointer(uintptr(c) + (hash&m)*uintptr(t.Bucketsize)))
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
			if b.tophash[i] != top {
				continue
			}
			k := runtimer.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.Keysize))
			if t.Indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if alg.Equal(key, k) {
				v := runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.Keysize)+i*uintptr(t.Valuesize))
				if t.Indirectvalue {
					v = *((*unsafe.Pointer)(v))
				}
				return v, true
			}
		}
		b = b.overflow(t)
		if b == nil {
			return unsafe.Pointer(&zeroVal[0]), false
		}
	}
}

// returns both key and value. Used by map iterator
func mapaccessK(t *runtimer.MapType, h *hmap, key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer) {
	if h == nil || h.count == 0 {
		return nil, nil
	}
	alg := t.Key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))
	m := uintptr(1)<<h.B - 1
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + (hash&m)*uintptr(t.Bucketsize)))
	if c := h.oldbuckets; c != nil {
		if !h.sameSizeGrow() {
			// There used to be half as many buckets; mask down one more power of two.
			m >>= 1
		}
		oldb := (*bmap)(unsafe.Pointer(uintptr(c) + (hash&m)*uintptr(t.Bucketsize)))
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
			if b.tophash[i] != top {
				continue
			}
			k := runtimer.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.Keysize))
			if t.Indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if alg.Equal(key, k) {
				v := runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.Keysize)+i*uintptr(t.Valuesize))
				if t.Indirectvalue {
					v = *((*unsafe.Pointer)(v))
				}
				return k, v
			}
		}
		b = b.overflow(t)
		if b == nil {
			return nil, nil
		}
	}
}

func mapaccess1_fat(t *runtimer.MapType, h *hmap, key, zero unsafe.Pointer) unsafe.Pointer {
	v := mapaccess1(t, h, key)
	if v == unsafe.Pointer(&zeroVal[0]) {
		return zero
	}
	return v
}

func mapaccess2_fat(t *runtimer.MapType, h *hmap, key, zero unsafe.Pointer) (unsafe.Pointer, bool) {
	v := mapaccess1(t, h, key)
	if v == unsafe.Pointer(&zeroVal[0]) {
		return zero, false
	}
	return v, true
}

// Like mapaccess, but allocates a slot for the key if it is not present in the map.
func mapassign(t *runtimer.MapType, h *hmap, key unsafe.Pointer) unsafe.Pointer {
	if h == nil {
		panic("assignment to entry in nil map")
	}
	if runtimer.Msanenabled {
		runtimer.Msanread(key, t.Key.Size)
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map writes")
	}
	alg := t.Key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))

	// Set hashWriting after calling alg.Hash, since alg.Hash may panic,
	// in which case we have not actually done a write.
	h.flags |= hashWriting

	if h.buckets == nil {
		h.buckets = runtimer.Newarray((*runtimer.Type)((unsafe.Pointer)(t.Bucket)), 1)
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
			k := runtimer.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.Keysize))
			if t.Indirectkey {
				k = *((*unsafe.Pointer)(k))
			}
			if !alg.Equal(key, k) {
				continue
			}
			// already have a mapping for key. Update it.
			if t.Needkeyupdate {
				runtimer.Typedmemmove(runtimer.PtrToType(unsafe.Pointer(t.Key)), k, key)
			}
			val = runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.Keysize)+i*uintptr(t.Valuesize))
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
		val = runtimer.Add(insertk, bucketCnt*uintptr(t.Keysize))
	}

	// store new key/value at insert position
	if t.Indirectkey {
		kmem := runtimer.Newobject(runtimer.PtrToType(unsafe.Pointer(t.Key)))
		*(*unsafe.Pointer)(insertk) = kmem
		insertk = kmem
	}
	if t.Indirectvalue {
		vmem := runtimer.Newobject(runtimer.PtrToType(unsafe.Pointer(t.Elem)))
		*(*unsafe.Pointer)(val) = vmem
	}
	runtimer.Typedmemmove(runtimer.PtrToType(unsafe.Pointer(t.Key)), insertk, key)
	*inserti = top
	h.count++

done:
	if h.flags&hashWriting == 0 {
		runtimer.Throw("concurrent map writes")
	}
	h.flags &^= hashWriting
	if t.Indirectvalue {
		val = *((*unsafe.Pointer)(val))
	}
	return val
}

func mapdelete(t *runtimer.MapType, h *hmap, key unsafe.Pointer) {
	if runtimer.Msanenabled && h != nil {
		runtimer.Msanread(key, t.Key.Size)
	}
	if h == nil || h.count == 0 {
		return
	}
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map writes")
	}

	alg := t.Key.Alg
	hash := alg.Hash(key, uintptr(h.hash0))

	// Set hashWriting after calling alg.Hash, since alg.Hash may panic,
	// in which case we have not actually done a write (delete).
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
			k := runtimer.Add(unsafe.Pointer(b), dataOffset+i*uintptr(t.Keysize))
			k2 := k
			if t.Indirectkey {
				k2 = *((*unsafe.Pointer)(k2))
			}
			if !alg.Equal(key, k2) {
				continue
			}
			if t.Indirectkey {
				*(*unsafe.Pointer)(k) = nil
			} else {
				runtimer.Typedmemclr(runtimer.PtrToType(unsafe.Pointer(t.Key)), k)
			}
			v := unsafe.Pointer(uintptr(unsafe.Pointer(b)) + dataOffset + bucketCnt*uintptr(t.Keysize) + i*uintptr(t.Valuesize))
			if t.Indirectvalue {
				*(*unsafe.Pointer)(v) = nil
			} else {
				runtimer.Typedmemclr(runtimer.PtrToType(unsafe.Pointer(t.Elem)), v)
			}
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

func mapiterinit(t *runtimer.MapType, h *hmap, it *hiter) {
	// Clear pointer fields so garbage collector does not complain.
	it.key = nil
	it.value = nil
	it.t = nil
	it.h = nil
	it.buckets = nil
	it.bptr = nil
	it.overflow[0] = nil
	it.overflow[1] = nil

	if h == nil || h.count == 0 {
		it.key = nil
		it.value = nil
		return
	}

	if unsafe.Sizeof(hiter{})/runtimer.PtrSize != 12 {
		runtimer.Throw("hash_iter size incorrect") // see ../../cmd/internal/gc/reflect.go
	}
	it.t = t
	it.h = h

	// grab snapshot of bucket state
	it.B = h.B
	it.buckets = h.buckets
	if t.Bucket.Kind&runtimer.KindNoPointers != 0 {
		// Allocate the current slice and remember pointers to both current and old.
		// This preserves all relevant overflow buckets alive even if
		// the table grows and/or overflow buckets are added to the table
		// while we are iterating.
		h.createOverflow()
		it.overflow = *h.overflow
	}

	// decide where to start
	r := uintptr(runtimer.Fastrand())
	if h.B > 31-bucketCntBits {
		r += uintptr(runtimer.Fastrand()) << 31
	}
	it.startBucket = r & (uintptr(1)<<h.B - 1)
	it.offset = uint8(r >> h.B & (bucketCnt - 1))

	// iterator state
	it.bucket = it.startBucket
	it.wrapped = false
	it.bptr = nil

	// Remember we have an iterator.
	// Can run concurrently with another hash_iter_init().
	if old := h.flags; old&(iterator|oldIterator) != iterator|oldIterator {
		runtimer.Or8(&h.flags, iterator|oldIterator)
	}

	mapiternext(it)
}

func mapiternext(it *hiter) {
	h := it.h
	if h.flags&hashWriting != 0 {
		runtimer.Throw("concurrent map iteration and map write")
	}
	t := it.t
	bucket := it.bucket
	b := it.bptr
	i := it.i
	checkBucket := it.checkBucket
	alg := t.Key.Alg

next:
	if b == nil {
		if bucket == it.startBucket && it.wrapped {
			// end of iteration
			it.key = nil
			it.value = nil
			return
		}
		if h.growing() && it.B == h.B {
			// Iterator was started in the middle of a grow, and the grow isn't done yet.
			// If the bucket we're looking at hasn't been filled in yet (i.e. the old
			// bucket hasn't been evacuated) then we need to iterate through the old
			// bucket and only return the ones that will be migrated to this bucket.
			oldbucket := bucket & it.h.oldbucketmask()
			b = (*bmap)(runtimer.Add(h.oldbuckets, oldbucket*uintptr(t.Bucketsize)))
			if !evacuated(b) {
				checkBucket = bucket
			} else {
				b = (*bmap)(runtimer.Add(it.buckets, bucket*uintptr(t.Bucketsize)))
				checkBucket = noCheck
			}
		} else {
			b = (*bmap)(runtimer.Add(it.buckets, bucket*uintptr(t.Bucketsize)))
			checkBucket = noCheck
		}
		bucket++
		if bucket == uintptr(1)<<it.B {
			bucket = 0
			it.wrapped = true
		}
		i = 0
	}
	for ; i < bucketCnt; i++ {
		offi := (i + it.offset) & (bucketCnt - 1)
		k := runtimer.Add(unsafe.Pointer(b), dataOffset+uintptr(offi)*uintptr(t.Keysize))
		v := runtimer.Add(unsafe.Pointer(b), dataOffset+bucketCnt*uintptr(t.Keysize)+uintptr(offi)*uintptr(t.Valuesize))
		if b.tophash[offi] != empty && b.tophash[offi] != evacuatedEmpty {
			if checkBucket != noCheck && !h.sameSizeGrow() {
				// Special case: iterator was started during a grow to a larger size
				// and the grow is not done yet. We're working on a bucket whose
				// oldbucket has not been evacuated yet. Or at least, it wasn't
				// evacuated when we started the bucket. So we're iterating
				// through the oldbucket, skipping any keys that will go
				// to the other new bucket (each oldbucket expands to two
				// buckets during a grow).
				k2 := k
				if t.Indirectkey {
					k2 = *((*unsafe.Pointer)(k2))
				}
				if t.Reflexivekey || alg.Equal(k2, k2) {
					// If the item in the oldbucket is not destined for
					// the current new bucket in the iteration, skip it.
					hash := alg.Hash(k2, uintptr(h.hash0))
					if hash&(uintptr(1)<<it.B-1) != checkBucket {
						continue
					}
				} else {
					// Hash isn't repeatable if k != k (NaNs).  We need a
					// repeatable and randomish choice of which direction
					// to send NaNs during evacuation. We'll use the low
					// bit of tophash to decide which way NaNs go.
					// NOTE: this case is why we need two evacuate tophash
					// values, evacuatedX and evacuatedY, that differ in
					// their low bit.
					if checkBucket>>(it.B-1) != uintptr(b.tophash[offi]&1) {
						continue
					}
				}
			}
			if b.tophash[offi] != evacuatedX && b.tophash[offi] != evacuatedY {
				// this is the golden data, we can return it.
				if t.Indirectkey {
					k = *((*unsafe.Pointer)(k))
				}
				it.key = k
				if t.Indirectvalue {
					v = *((*unsafe.Pointer)(v))
				}
				it.value = v
			} else {
				// The hash table has grown since the iterator was started.
				// The golden data for this key is now somewhere else.
				k2 := k
				if t.Indirectkey {
					k2 = *((*unsafe.Pointer)(k2))
				}
				if t.Reflexivekey || alg.Equal(k2, k2) {
					// Check the current hash table for the data.
					// This code handles the case where the key
					// has been deleted, updated, or deleted and reinserted.
					// NOTE: we need to regrab the key as it has potentially been
					// updated to an equal() but not identical key (e.g. +0.0 vs -0.0).
					rk, rv := mapaccessK(t, h, k2)
					if rk == nil {
						continue // key has been deleted
					}
					it.key = rk
					it.value = rv
				} else {
					// if key!=key then the entry can't be deleted or
					// updated, so we can just return it. That's lucky for
					// us because when key!=key we can't look it up
					// successfully in the current table.
					it.key = k2
					if t.Indirectvalue {
						v = *((*unsafe.Pointer)(v))
					}
					it.value = v
				}
			}
			it.bucket = bucket
			if it.bptr != b { // avoid unnecessary write barrier; see issue 14921
				it.bptr = b
			}
			it.i = i + 1
			it.checkBucket = checkBucket
			return
		}
	}
	b = b.overflow(t)
	i = 0
	goto next
}

func hashGrow(t *runtimer.MapType, h *hmap) {
	// If we've hit the load factor, get bigger.
	// Otherwise, there are too many overflow buckets,
	// so keep the same number of buckets and "grow" laterally.
	bigger := uint8(1)
	if !overLoadFactor(int64(h.count), h.B) {
		bigger = 0
		h.flags |= sameSizeGrow
	}
	oldbuckets := h.buckets
	newbuckets := runtimer.Newarray(runtimer.PtrToType(unsafe.Pointer(t.Bucket)), 1<<(h.B+bigger))
	flags := h.flags &^ (iterator | oldIterator)
	if h.flags&iterator != 0 {
		flags |= oldIterator
	}
	// commit the grow (atomic wrt gc)
	h.B += bigger
	h.flags = flags
	h.oldbuckets = oldbuckets
	h.buckets = newbuckets
	h.nevacuate = 0
	h.noverflow = 0

	if h.overflow != nil {
		// Promote current overflow buckets to the old generation.
		if h.overflow[1] != nil {
			runtimer.Throw("overflow is not nil")
		}
		h.overflow[1] = h.overflow[0]
		h.overflow[0] = nil
	}

	// the actual copying of the hash table data is done incrementally
	// by growWork() and evacuate().
}

// overLoadFactor reports whether count items placed in 1<<B buckets is over loadFactor.
func overLoadFactor(count int64, B uint8) bool {
	// TODO: rewrite to use integer math and comparison?
	return count >= bucketCnt && float32(count) >= loadFactor*float32((uintptr(1)<<B))
}

// tooManyOverflowBuckets reports whether noverflow buckets is too many for a map with 1<<B buckets.
// Note that most of these overflow buckets must be in sparse use;
// if use was dense, then we'd have already triggered regular map growth.
func tooManyOverflowBuckets(noverflow uint16, B uint8) bool {
	// If the threshold is too low, we do extraneous work.
	// If the threshold is too high, maps that grow and shrink can hold on to lots of unused memory.
	// "too many" means (approximately) as many overflow buckets as regular buckets.
	// See incrnoverflow for more details.
	if B < 16 {
		return noverflow >= uint16(1)<<B
	}
	return noverflow >= 1<<15
}

// growing reports whether h is growing. The growth may be to the same size or bigger.
func (h *hmap) growing() bool {
	return h.oldbuckets != nil
}

// sameSizeGrow reports whether the current growth is to a map of the same size.
func (h *hmap) sameSizeGrow() bool {
	return h.flags&sameSizeGrow != 0
}

// noldbuckets calculates the number of buckets prior to the current map growth.
func (h *hmap) noldbuckets() uintptr {
	oldB := h.B
	if !h.sameSizeGrow() {
		oldB--
	}
	return uintptr(1) << oldB
}

// oldbucketmask provides a mask that can be applied to calculate n % noldbuckets().
func (h *hmap) oldbucketmask() uintptr {
	return h.noldbuckets() - 1
}

func growWork(t *runtimer.MapType, h *hmap, bucket uintptr) {
	// make sure we evacuate the oldbucket corresponding
	// to the bucket we're about to use
	evacuate(t, h, bucket&h.oldbucketmask())

	// evacuate one more oldbucket to make progress on growing
	if h.growing() {
		evacuate(t, h, h.nevacuate)
	}
}

func bucketEvacuated(t *runtimer.MapType, h *hmap, bucket uintptr) bool {
	b := (*bmap)(runtimer.Add(h.oldbuckets, bucket*uintptr(t.Bucketsize)))
	return evacuated(b)
}

func evacuate(t *runtimer.MapType, h *hmap, oldbucket uintptr) {
	b := (*bmap)(runtimer.Add(h.oldbuckets, oldbucket*uintptr(t.Bucketsize)))
	newbit := h.noldbuckets()
	alg := t.Key.Alg
	if !evacuated(b) {
		// TODO: reuse overflow buckets instead of using new ones, if there
		// is no iterator using the old buckets.  (If !oldIterator.)

		var (
			x, y   *bmap          // current low/high buckets in new map
			xi, yi int            // key/val indices into x and y
			xk, yk unsafe.Pointer // pointers to current x and y key storage
			xv, yv unsafe.Pointer // pointers to current x and y value storage
		)
		x = (*bmap)(runtimer.Add(h.buckets, oldbucket*uintptr(t.Bucketsize)))
		xi = 0
		xk = runtimer.Add(unsafe.Pointer(x), dataOffset)
		xv = runtimer.Add(xk, bucketCnt*uintptr(t.Keysize))
		if !h.sameSizeGrow() {
			// Only calculate y pointers if we're growing bigger.
			// Otherwise GC can see bad pointers.
			y = (*bmap)(runtimer.Add(h.buckets, (oldbucket+newbit)*uintptr(t.Bucketsize)))
			yi = 0
			yk = runtimer.Add(unsafe.Pointer(y), dataOffset)
			yv = runtimer.Add(yk, bucketCnt*uintptr(t.Keysize))
		}
		for ; b != nil; b = b.overflow(t) {
			k := runtimer.Add(unsafe.Pointer(b), dataOffset)
			v := runtimer.Add(k, bucketCnt*uintptr(t.Keysize))
			for i := 0; i < bucketCnt; i, k, v = i+1, runtimer.Add(k, uintptr(t.Keysize)), runtimer.Add(v, uintptr(t.Valuesize)) {
				top := b.tophash[i]
				if top == empty {
					b.tophash[i] = evacuatedEmpty
					continue
				}
				if top < minTopHash {
					runtimer.Throw("bad map state")
				}
				k2 := k
				if t.Indirectkey {
					k2 = *((*unsafe.Pointer)(k2))
				}
				useX := true
				if !h.sameSizeGrow() {
					// Compute hash to make our evacuation decision (whether we need
					// to send this key/value to bucket x or bucket y).
					hash := alg.Hash(k2, uintptr(h.hash0))
					if h.flags&iterator != 0 {
						if !t.Reflexivekey && !alg.Equal(k2, k2) {
							// If key != key (NaNs), then the hash could be (and probably
							// will be) entirely different from the old hash. Moreover,
							// it isn't reproducible. Reproducibility is required in the
							// presence of iterators, as our evacuation decision must
							// match whatever decision the iterator made.
							// Fortunately, we have the freedom to send these keys either
							// way. Also, tophash is meaningless for these kinds of keys.
							// We let the low bit of tophash drive the evacuation decision.
							// We recompute a new random tophash for the next level so
							// these keys will get evenly distributed across all buckets
							// after multiple grows.
							if top&1 != 0 {
								hash |= newbit
							} else {
								hash &^= newbit
							}
							top = uint8(hash >> (runtimer.PtrSize*8 - 8))
							if top < minTopHash {
								top += minTopHash
							}
						}
					}
					useX = hash&newbit == 0
				}
				if useX {
					b.tophash[i] = evacuatedX
					if xi == bucketCnt {
						newx := (*bmap)(runtimer.Newobject(runtimer.PtrToType(unsafe.Pointer(t.Bucket))))
						h.setoverflow(t, x, newx)
						x = newx
						xi = 0
						xk = runtimer.Add(unsafe.Pointer(x), dataOffset)
						xv = runtimer.Add(xk, bucketCnt*uintptr(t.Keysize))
					}
					x.tophash[xi] = top
					if t.Indirectkey {
						*(*unsafe.Pointer)(xk) = k2 // copy pointer
					} else {
						runtimer.Typedmemmove(runtimer.PtrToType(unsafe.Pointer(t.Key)), xk, k) // copy value
					}
					if t.Indirectvalue {
						*(*unsafe.Pointer)(xv) = *(*unsafe.Pointer)(v)
					} else {
						runtimer.Typedmemmove(runtimer.PtrToType(unsafe.Pointer(t.Elem)), xv, v)
					}
					xi++
					xk = runtimer.Add(xk, uintptr(t.Keysize))
					xv = runtimer.Add(xv, uintptr(t.Valuesize))
				} else {
					b.tophash[i] = evacuatedY
					if yi == bucketCnt {
						newy := (*bmap)(runtimer.Newobject(runtimer.PtrToType(unsafe.Pointer(t.Bucket))))
						h.setoverflow(t, y, newy)
						y = newy
						yi = 0
						yk = runtimer.Add(unsafe.Pointer(y), dataOffset)
						yv = runtimer.Add(yk, bucketCnt*uintptr(t.Keysize))
					}
					y.tophash[yi] = top
					if t.Indirectkey {
						*(*unsafe.Pointer)(yk) = k2
					} else {
						runtimer.Typedmemmove(runtimer.PtrToType(unsafe.Pointer(t.Key)), yk, k)
					}
					if t.Indirectvalue {
						*(*unsafe.Pointer)(yv) = *(*unsafe.Pointer)(v)
					} else {
						runtimer.Typedmemmove(runtimer.PtrToType(unsafe.Pointer(t.Elem)), yv, v)
					}
					yi++
					yk = runtimer.Add(yk, uintptr(t.Keysize))
					yv = runtimer.Add(yv, uintptr(t.Valuesize))
				}
			}
		}
		// Unlink the overflow buckets & clear key/value to help GC.
		if h.flags&oldIterator == 0 {
			b = (*bmap)(runtimer.Add(h.oldbuckets, oldbucket*uintptr(t.Bucketsize)))
			// Preserve b.tophash because the evacuation
			// state is maintained there.
			if t.Bucket.Kind&runtimer.KindNoPointers == 0 {
				runtimer.MemclrHasPointers(runtimer.Add(unsafe.Pointer(b), dataOffset), uintptr(t.Bucketsize)-dataOffset)
			} else {
				runtimer.MemclrNoHeapPointers(runtimer.Add(unsafe.Pointer(b), dataOffset), uintptr(t.Bucketsize)-dataOffset)
			}
		}

	}

	// Advance evacuation mark
	if oldbucket == h.nevacuate {
		h.nevacuate = oldbucket + 1
		// Experiments suggest that 1024 is overkill by at least an order of magnitude.
		// Put it in there as a safeguard anyway, to ensure O(1) behavior.
		stop := h.nevacuate + 1024
		if stop > newbit {
			stop = newbit
		}
		for h.nevacuate != stop && bucketEvacuated(t, h, h.nevacuate) {
			h.nevacuate++
		}
		if h.nevacuate == newbit { // newbit == # of oldbuckets
			// Growing is all done. Free old main bucket array.
			h.oldbuckets = nil
			// Can discard old overflow buckets as well.
			// If they are still referenced by an iterator,
			// then the iterator holds a pointers to the slice.
			if h.overflow != nil {
				h.overflow[1] = nil
			}
			h.flags &^= sameSizeGrow
		}
	}
}

func ismapkey(t *runtimer.Type) bool {
	return t.Alg.Hash != nil
}

const maxZero = 1024 // must match value in ../cmd/compile/internal/gc/walk.go
var zeroVal [maxZero]byte
