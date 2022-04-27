package group

import (
	"unsafe"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
)

func Group[T any](vecs []any, rows int) error {
	keys := make([]uint64, UnitLimit)
	keyOffs := make([]uint32, UnitLimit)
	zKeys := make([]uint64, UnitLimit)
	zKeyOffs := make([]uint32, UnitLimit)
	count := rows
	for i := 0; i < count; i += UnitLimit {
		n := count - i
		if n > UnitLimit {
			n = UnitLimit
		}
		copy(keys, zKeys)
		copy(keyOffs, zKeyOffs)
		for _, vec := range vecs {
			switch v := (interface{})(vec).(type) {
			case *vector.Vector[types.Int8]:
				fillGroup(v, n, uint32(types.TypeSize(v.Typ.Oid)), keys, keyOffs)
			case *vector.Vector[types.Int16]:
				fillGroup(v, n, uint32(types.TypeSize(v.Typ.Oid)), keys, keyOffs)
			case *vector.Vector[types.Int32]:
				fillGroup(v, n, uint32(types.TypeSize(v.Typ.Oid)), keys, keyOffs)
			case *vector.Vector[types.Int64]:
				fillGroup(v, n, uint32(types.TypeSize(v.Typ.Oid)), keys, keyOffs)
			case *vector.Vector[types.Bytes]:
				for i := 0; i < n; i++ {
					copy(unsafe.Slice((*byte)(unsafe.Pointer(&keys[i])), 8)[keyOffs[i]:], v.Col[i])
					keyOffs[i] += uint32(len(v.Col[i]))
				}
			}
		}
	}
	return nil
}

func fillGroup[T types.Element[T]](vec *vector.Vector[T], n int, sz uint32, keys []uint64, keyOffs []uint32) error {
	for i := 0; i < n; i++ {
		*(*T)(unsafe.Add(unsafe.Pointer(&keys[i]), keyOffs[i])) = vec.Col[i]
		keyOffs[i] += sz
	}
	return nil
}
