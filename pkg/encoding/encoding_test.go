package encoding

import (
	"fmt"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/container/types"
)

func TestEncode(t *testing.T) {
	vs := make([]types.Int64, 10)
	for i := 0; i < 10; i++ {
		vs[i] = types.Int64(i)
	}
	data := EncodeFixedSlice(vs, 8)
	fmt.Printf("data: %v\n", data)
	rs := DecodeFixedSlice[types.Int64](data, 8)
	fmt.Printf("rs: %v\n", rs)
}
