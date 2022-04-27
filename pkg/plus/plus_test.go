package plus

import (
	"math/rand"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/plus/int64s"
)

const Loop = 1000

var Xs, Ys, Zs []int64
var Vx, Vy, Vz *vector.Vector[types.Int64]

func init() {
	Xs = make([]int64, Loop)
	Ys = make([]int64, Loop)
	Zs = make([]int64, Loop)
	Vx = vector.New[types.Int64](types.New(types.T_int64))
	Vy = vector.New[types.Int64](types.New(types.T_int64))
	Vz = vector.New[types.Int64](types.New(types.T_int64))
	Vz.Col = make([]types.Int64, Loop)
	for i := 0; i < Loop; i++ {
		x := rand.Intn(Loop * 10)
		y := rand.Intn(Loop * 10)
		Xs[i] = int64(x)
		Ys[i] = int64(y)
		Vx.Col = append(Vx.Col, types.Int64(x))
		Vy.Col = append(Vy.Col, types.Int64(x))
	}

}

func BenchmarkPlus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		int64s.Plus(Xs, Ys, Zs)
	}
}

func BenchmarkGenericPlus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		VectorPlus(Vx, Vy, Vz)
	}
}
