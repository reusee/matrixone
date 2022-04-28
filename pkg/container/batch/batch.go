package batch

import "github.com/matrixorigin/matrixone/pkg/container/vector"

func New(attrs []string) *Batch {
	return &Batch{
		Attrs: attrs,
		Vecs:  make([]vector.AnyVector, len(attrs)),
	}
}

func (b *Batch) SetLength(n int) {
	for _, vec := range b.Vecs {
		vec.SetLength(n)
	}
}
