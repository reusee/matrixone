package int64s

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
)

const (
	UnitLimit = 1204
)

type Vector struct {
	Typ types.Type
	Col interface{}
}
