package group

import (
	"unsafe"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
)

func Group(vecs []vector.AnyVector, rows int) error {
	keys := make([]uint64, UnitLimit)
	zKeys := make([]uint64, UnitLimit)
	keyOffs := make([]uint32, UnitLimit)
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
			switch typ := vec.Type(); typ.Oid {
			case types.T_int8:
				fillGroup((any)(vec).(*vector.Vector[types.Int8]), n, uint32(types.TypeSize(typ.Oid)), keys, keyOffs)
			case types.T_int16:
				fillGroup((any)(vec).(*vector.Vector[types.Int16]), n, uint32(types.TypeSize(typ.Oid)), keys, keyOffs)
			case types.T_int32:
				fillGroup((any)(vec).(*vector.Vector[types.Int32]), n, uint32(types.TypeSize(typ.Oid)), keys, keyOffs)
			case types.T_int64:
				fillGroup((any)(vec).(*vector.Vector[types.Int64]), n, uint32(types.TypeSize(typ.Oid)), keys, keyOffs)
			case types.T_char, types.T_varchar:
				v := (any)(vec).(*vector.Vector[types.Bytes])
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
