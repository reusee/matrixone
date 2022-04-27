package group

import (
	"strconv"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
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
	vecs := make([]vector.VectorLike, 2)
	vx := vector.New[types.Int64](types.New(types.T_int64))
	for i := 0; i < Loop; i++ {
		vx.Append(types.Int64(i), m)
	}
	vy := vector.New[types.Bytes](types.New(types.T_varchar))
	for i := 0; i < Loop; i++ {
		vy.Append(types.Bytes(strconv.Itoa(i)), m)
	}
	vecs[0] = vx
	vecs[1] = vx
	Group(vecs, Loop)

}
