package vector

import (
	"fmt"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/vm/mheap"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/guest"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/host"
)

func TestAppend(t *testing.T) {
	hm := host.New(1 << 20)
	gm := guest.New(1<<20, hm)
	m := mheap.New(gm)
	vx := New[types.Int64](types.New(types.T_int64))
	vx.Append(types.Int64(1), m)
	vx.Append(types.Int64(2), m)
	vx.Append(types.Int64(3), m)
	fmt.Printf("vx: %v: %v\n", vx.Col, vx.Data)
	vx.Reset()
	vx.Append(types.Int64(3), m)
	vx.Append(types.Int64(1), m)
	vx.Append(types.Int64(2), m)
	fmt.Printf("vx: %v: %v\n", vx.Col, vx.Data)
}

func TestAppendStr(t *testing.T) {
	hm := host.New(1 << 20)
	gm := guest.New(1<<20, hm)
	m := mheap.New(gm)
	vx := New[types.Bytes](types.New(types.T_varchar))
	vx.Append(types.Bytes("1"), m)
	vx.Append(types.Bytes("2"), m)
	vx.Append(types.Bytes("3"), m)
	fmt.Printf("vx: %v: %v\n", vx.Col, vx.Data)
	vx.Reset()
	vx.Append(types.Bytes("3"), m)
	vx.Append(types.Bytes("1"), m)
	vx.Append(types.Bytes("2"), m)
	fmt.Printf("vx: %v: %v\n", vx.Col, vx.Data)
}
