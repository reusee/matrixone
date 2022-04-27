package plus

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"golang.org/x/exp/constraints"
)

func VectorPlus[T types.Element[T]](vx, vy, vz *vector.Vector[T]) {
	switch vs := (interface{})(vx.Col).(type) {
	case []types.Int8:
		Plus(vs, (interface{})(vy.Col).([]types.Int8), (interface{})(vz.Col).([]types.Int8))
	case []types.Int16:
		Plus(vs, (interface{})(vy.Col).([]types.Int16), (interface{})(vz.Col).([]types.Int16))
	case []types.Int32:
		Plus(vs, (interface{})(vy.Col).([]types.Int32), (interface{})(vz.Col).([]types.Int32))
	case []types.Int64:
		Plus(vs, (interface{})(vy.Col).([]types.Int64), (interface{})(vz.Col).([]types.Int64))
	}

}

func Plus[T constraints.Integer | constraints.Float](xs, ys, zs []T) []T {
	for i, x := range xs {
		zs[i] = x + ys[i]
	}
	return zs
}
