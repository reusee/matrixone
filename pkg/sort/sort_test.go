package sort

import (
	"math/rand"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/sort/int64s"
)

const Loop = 1000

var Xs []int64
var Vec *vector.Vector[types.Int64]

func init() {
	Xs = make([]int64, Loop)
	Vec = vector.New[types.Int64](types.New(types.T_int64))
	for i := 0; i < Loop; i++ {
		x := rand.Intn(Loop * 10)
		Xs[i] = int64(x)
		Vec.Col = append(Vec.Col, types.Int64(x))
	}
}

func BenchmarkSort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		int64s.Sort(Xs)
	}
}

func BenchmarkGenericSort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		VectorSort(Vec)
	}
}
