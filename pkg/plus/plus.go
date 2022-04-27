package plus

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"golang.org/x/exp/constraints"
)

func VectorPlus(vx, vy, vz vector.VectorLike) {
	switch xs := (interface{})(vx).(type) {
	case *vector.Vector[types.Int8]:
		Plus(xs.Col, (interface{})(vy).(*vector.Vector[types.Int8]).Col, (interface{})(vz).(*vector.Vector[types.Int8]).Col)
	case *vector.Vector[types.Int16]:
		Plus(xs.Col, (interface{})(vy).(*vector.Vector[types.Int16]).Col, (interface{})(vz).(*vector.Vector[types.Int16]).Col)
	case *vector.Vector[types.Int32]:
		Plus(xs.Col, (interface{})(vy).(*vector.Vector[types.Int32]).Col, (interface{})(vz).(*vector.Vector[types.Int32]).Col)
	case *vector.Vector[types.Int64]:
		Plus(xs.Col, (interface{})(vy).(*vector.Vector[types.Int64]).Col, (interface{})(vz).(*vector.Vector[types.Int64]).Col)
	}
}

func Plus[T constraints.Integer | constraints.Float](xs, ys, zs []T) []T {
	for i, x := range xs {
		zs[i] = x + ys[i]
	}
	return zs
}
