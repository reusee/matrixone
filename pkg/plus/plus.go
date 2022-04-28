package plus

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"golang.org/x/exp/constraints"
)

func VectorPlus(vx, vy, vz vector.AnyVector) {
	switch vx.Type().Oid {
	case types.T_int8:
		Plus((any)(vx).(*vector.Vector[types.Int8]).Col, (any)(vy).(*vector.Vector[types.Int8]).Col, (any)(vz).(*vector.Vector[types.Int8]).Col)
	case types.T_int16:
		Plus((any)(vx).(*vector.Vector[types.Int16]).Col, (any)(vy).(*vector.Vector[types.Int16]).Col, (any)(vz).(*vector.Vector[types.Int16]).Col)
	case types.T_int32:
		Plus((any)(vx).(*vector.Vector[types.Int32]).Col, (any)(vy).(*vector.Vector[types.Int32]).Col, (any)(vz).(*vector.Vector[types.Int32]).Col)
	case types.T_int64:
		Plus((any)(vx).(*vector.Vector[types.Int64]).Col, (any)(vy).(*vector.Vector[types.Int64]).Col, (any)(vz).(*vector.Vector[types.Int64]).Col)
	}
}

func Plus[T constraints.Integer | constraints.Float](xs, ys, zs []T) []T {
	for i, x := range xs {
		zs[i] = x + ys[i]
	}
	return zs
}
