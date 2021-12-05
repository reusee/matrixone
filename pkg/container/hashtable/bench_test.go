package hashtable

import (
	"testing"
	"unsafe"
)

func BenchmarkInt64HashMapInsert(b *testing.B) {
	m := new(Int64HashMap)
	m.Init()
	n := 42
	ptr := unsafe.Pointer(&n)
	for i := 0; i < b.N; i++ {
		m.Insert(0, ptr)
	}
}

func BenchmarkInt64HashMapFind(b *testing.B) {
	m := new(Int64HashMap)
	m.Init()
	n := 42
	ptr := unsafe.Pointer(&n)
	m.Insert(0, ptr)
	for i := 0; i < b.N; i++ {
		m.Find(0, ptr)
	}
}
