package int64s

import (
	"unsafe"

	"github.com/matrixorigin/matrixone/pkg/container/types"
)

func Group(vecs []Vector, rows int) error {
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
			switch vec.Typ.Oid {
			case types.T_int64:
				fillGroup(vec.Col.([]int64), n, uint32(8), keys, keyOffs)
			}
		}
	}
	return nil

}

func fillGroup(vec []int64, n int, sz uint32, keys []uint64, keyOffs []uint32) error {
	for i := 0; i < n; i++ {
		*(*int64)(unsafe.Add(unsafe.Pointer(&keys[i]), keyOffs[i])) = vec[i]
		keyOffs[i] += sz
	}
	return nil
}
