// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vector

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/encoding"
	"github.com/matrixorigin/matrixone/pkg/vm/mheap"
)

func New[T types.Element[T]](typ types.Type) *Vector[T] {
	return &Vector[T]{
		Typ:  typ,
		Col:  []T{},
		Data: []byte{},
	}
}

func Reset[T types.Element[T]](v *Vector[T]) {
	v.Col = v.Col[:0]
	v.Data = v.Data[:0]
	if len(v.Offsets) > 0 {
		v.Offsets = v.Offsets[:0]
		v.Lengths = v.Lengths[:0]
	}
}

func Free[T types.Element[T]](v *Vector[T], m *mheap.Mheap) {
	mheap.Free(m, v.Data)
}

func Realloc[T types.Element[T]](v *Vector[T], size int, m *mheap.Mheap) error {
	oldLen := len(v.Data)
	data, err := mheap.Grow(m, v.Data, int64(oldLen+size))
	if err != nil {
		return err
	}
	mheap.Free(m, v.Data)
	v.Data = data[:oldLen]
	switch vec := (interface{})(v).(type) {
	case *Vector[types.Bytes]:
		vec.Col = vec.Col[:0]
		for i, off := range vec.Offsets {
			vec.Col = append(vec.Col, vec.Data[off:off+vec.Lengths[i]])
		}
	default:
		v.Col = encoding.DecodeFixedSlice[T](v.Data[:len(data)], size)[:oldLen/size]
	}
	return nil
}

func Append[T types.Element[T]](v *Vector[T], w T, m *mheap.Mheap) error {
	n := len(v.Col)
	if n+1 >= cap(v.Col) {
		if err := Realloc(v, w.Size(), m); err != nil {
			return err
		}
	}
	switch vec := (interface{})(v).(type) {
	case *Vector[types.Bytes]:
		wv, _ := (interface{})(w).(types.Bytes)
		vec.Lengths = append(vec.Lengths, uint64(len(wv)))
		vec.Offsets = append(vec.Offsets, uint64(len(v.Data)))
		size := len(vec.Data)
		vec.Data = append(vec.Data, wv...)
		vec.Col = append(vec.Col, vec.Data[size:size+len(wv)])
	default:
		v.Col = append(v.Col, w)
		v.Data = v.Data[:(n+1)*w.Size()]
	}
	return nil
}
