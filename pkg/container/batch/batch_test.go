package batch

import (
	"fmt"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
)

func TestBacth(t *testing.T) {
	bat := New([]string{"id", "price"})
	bat.Vecs[0] = vector.New[types.Int64](types.New(types.T_int64))
	bat.Vecs[1] = vector.New[types.Bytes](types.New(types.T_varchar))
	bat.SetLength(0)
	fmt.Printf("%v\n", bat)
}
