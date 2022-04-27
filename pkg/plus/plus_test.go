package plus

import (
	"math/rand"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/plus/int64s"
	"github.com/matrixorigin/matrixone/pkg/vm/mheap"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/guest"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/host"
)

const Loop = 1000

var Xs, Ys, Zs []int64
var Vx, Vy, Vz *vector.Vector[types.Int64]

func init() {
	hm := host.New(1 << 20)
	gm := guest.New(1<<20, hm)
	m := mheap.New(gm)
	Xs = make([]int64, Loop)
	Ys = make([]int64, Loop)
	Zs = make([]int64, Loop)
	Vx = vector.New[types.Int64](types.New(types.T_int64))
	Vy = vector.New[types.Int64](types.New(types.T_int64))
	Vz = vector.New[types.Int64](types.New(types.T_int64))
	for i := 0; i < Loop; i++ {
		x := rand.Intn(Loop * 10)
		y := rand.Intn(Loop * 10)
		Xs[i] = int64(x)
		Ys[i] = int64(y)
		Vx.Append(types.Int64(x), m)
		Vy.Append(types.Int64(x), m)
		Vz.Append(types.Int64(0), m)
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
