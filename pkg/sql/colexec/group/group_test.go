package group

import (
	"strconv"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec/group/int64s"
	"github.com/matrixorigin/matrixone/pkg/vm/mheap"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/guest"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/host"
)

const (
	Loop = 10000
)

func TestGroup(t *testing.T) {
	hm := host.New(1 << 20)
	gm := guest.New(1<<20, hm)
	m := mheap.New(gm)
	vecs := make([]vector.AnyVector, 2)
	{
		vec := vector.New[types.Int64](types.New(types.T_int64))
		for i := 0; i < Loop; i++ {
			vec.Append(types.Int64(i), m)
		}
		vecs[0] = vec
	}
	{
		vec := vector.New[types.Bytes](types.New(types.T_varchar))
		for i := 0; i < Loop; i++ {
			vec.Append(types.Bytes(strconv.Itoa(i)), m)
		}
		vecs[1] = vec
	}
	Group(vecs, Loop)
}

var Vecs []vector.AnyVector
var Int64Vecs []int64s.Vector

func init() {
	hm := host.New(1 << 20)
	gm := guest.New(1<<20, hm)
	m := mheap.New(gm)
	Vecs = make([]vector.AnyVector, 2)
	Int64Vecs = make([]int64s.Vector, 2)
	{
		vs := make([]int64, Loop)
		vec := vector.New[types.Int64](types.New(types.T_int64))
		for i := 0; i < Loop; i++ {
			vs[i] = int64(i)
			vec.Append(types.Int64(i), m)
		}
		Vecs[0] = vec
		Int64Vecs[0] = int64s.Vector{
			Col: vs,
			Typ: types.New(types.T_int64),
		}
	}
	{
		vs := make([]int64, Loop)
		vec := vector.New[types.Int64](types.New(types.T_int64))
		for i := 0; i < Loop; i++ {
			vs[i] = int64(i)
			vec.Append(types.Int64(i), m)
		}
		Vecs[1] = vec
		Int64Vecs[1] = int64s.Vector{
			Col: vs,
			Typ: types.New(types.T_int64),
		}

	}
}

func BenchmarkGroup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		int64s.Group(Int64Vecs, Loop)
	}
}

func BenchmarkGenericGroup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Group(Vecs, Loop)
	}
}
