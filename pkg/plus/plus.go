package plus

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"golang.org/x/exp/constraints"
)

type NumberElement[T any] interface {
	types.Element[T]
	constraints.Integer | constraints.Float
}

func VectorPlus[T NumberElement[T]](xv, yv, zv *vector.Vector[T]) []T {
	for i, x := range xv.Col {
		zv.Col[i] = x + yv.Col[i]
	}
	return zv.Col
}
