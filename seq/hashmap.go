package seq

import (
	"fmt"

	"github.com/mediocregopher/ginger/types"
)

// Hash maps are built on top of hash sets. KeyVal implements Setable, but the
// Hash and Equal methods only apply to the key and ignore the value.

// Container for a key/value pair, used by HashMap to hold its data
type KV struct {
	Key types.Elem
	Val types.Elem
}

func KeyVal(key, val types.Elem) *KV {
	return &KV{key, val}
}

// Implementation of Hash for Setable. Only actually hashes the Key field
func (kv *KV) Hash(i uint32) uint32 {
	return hash(kv.Key, i)
}

// Implementation of Equal for Setable. Only actually compares the key field. If
// compared to another KV, only compares the other key as well.
func (kv *KV) Equal(v types.Elem) bool {
	if kv2, ok := v.(*KV); ok {
		return kv.Key.Equal(kv2.Key)
	}
	return kv.Key.Equal(v)
}

func (kv *KV) fullEqual(v types.Elem) bool {
	kv2, ok := v.(*KV)
	if !ok {
		return false
	}

	return kv.Key.Equal(kv2.Key) && kv.Val.Equal(kv2.Val)
}

// Implementation of String for Stringer
func (kv *KV) String() string {
	return fmt.Sprintf("%v -> %v", kv.Key, kv.Val)
}

// HashMaps are actually built on top of Sets, just with some added convenience
// methods for interacting with them as actual key/val stores
type HashMap struct {
	set *Set
}

// Returns a new HashMap of the given KVs (or possibly just an empty HashMap)
func NewHashMap(kvs ...*KV) *HashMap {
	ints := make([]types.Elem, len(kvs))
	for i := range kvs {
		ints[i] = kvs[i]
	}
	return &HashMap{
		set: NewSet(ints...),
	}
}

// Implementation of FirstRest for Seq interface. First return value will
// always be a *KV or nil. Completes in O(log(N)) time.
func (hm *HashMap) FirstRest() (types.Elem, Seq, bool) {
	if hm == nil {
		return nil, nil, false
	}
	el, nset, ok := hm.set.FirstRest()
	return el, &HashMap{nset.(*Set)}, ok
}

// Implementation of Equal for types.Elem interface. Completes in O(Nlog(M))
// time if e is another HashMap, where M is the size of the given HashMap
func (hm *HashMap) Equal(e types.Elem) bool {
	// This can't just use Set's Equal because that would end up using KeyVal's
	// Equal, which is not a true Equal

	hm2, ok := e.(*HashMap)
	if !ok {
		return false
	}

	var el types.Elem
	s := Seq(hm)
	size := uint64(0)

	for { 
		el, s, ok = s.FirstRest()
		if !ok {
			return size == hm2.Size()
		}
		size++

		kv := el.(*KV)
		k, v := kv.Key, kv.Val

		v2, ok := hm2.Get(k)
		if !ok || !v.Equal(v2) {
			return false
		}
	}
}

// Returns a new HashMap with the given value set on the given key. Also returns
// whether or not this was the first time setting that key (false if it was
// already there and was overwritten). Has the same complexity as Set's SetVal
// method.
func (hm *HashMap) Set(key, val types.Elem) (*HashMap, bool) {
	if hm == nil {
		hm = NewHashMap()
	}

	nset, ok := hm.set.SetVal(KeyVal(key, val))
	return &HashMap{nset}, ok
}

// Returns a new HashMap with the given key removed from it. Also returns
// whether or not the key was already there (true if so, false if not). Has the
// same time complexity as Set's DelVal method.
func (hm *HashMap) Del(key types.Elem) (*HashMap, bool) {
	if hm == nil {
		hm = NewHashMap()
	}

	nset, ok := hm.set.DelVal(KeyVal(key, nil))
	return &HashMap{nset}, ok
}

// Returns a value for a given key from the HashMap, along with a boolean
// indicating whether or not the value was found. Has the same time complexity
// as Set's GetVal method.
func (hm *HashMap) Get(key types.Elem) (types.Elem, bool) {
	if hm == nil {
		return nil, false
	} else if kv, ok := hm.set.GetVal(KeyVal(key, nil)); ok {
		return kv.(*KV).Val, true
	} else {
		return nil, false
	}
}

// Same as FirstRest, but returns values already casted, which may be convenient
// in some cases.
func (hm *HashMap) FirstRestKV() (*KV, *HashMap, bool) {
	if el, nhm, ok := hm.FirstRest(); ok {
		return el.(*KV), nhm.(*HashMap), true
	} else {
		return nil, nil, false
	}
}

// Implementation of String for Stringer interface
func (hm *HashMap) String() string {
	return ToString(hm, "{", "}")
}

// Returns the number of KVs in the HashMap. Has the same complexity as Set's
// Size method.
func (hm *HashMap) Size() uint64 {
	return hm.set.Size()
}
