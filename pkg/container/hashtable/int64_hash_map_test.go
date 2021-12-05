package hashtable

import (
	"testing"
	"unsafe"
)

func TestInt64Hashmap(t *testing.T) {
	m := new(Int64HashMap)
	m.Init()
	const n = 100_0000
	for i := uint64(0); i < n; i++ {
		ptr := unsafe.Pointer(&i)
		inserted, valuePtr := m.Insert(0, ptr)
		if !inserted {
			t.Fatal()
		}
		*valuePtr = i
	}
	for i := uint64(0); i < n; i++ {
		ptr := unsafe.Pointer(&i)
		valuePtr := m.Find(0, ptr)
		if valuePtr == nil {
			t.Fatal()
		}
		if *valuePtr != i {
			t.Fatal()
		}
	}
}
