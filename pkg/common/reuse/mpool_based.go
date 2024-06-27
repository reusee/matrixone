// Copyright 2023 Matrix Origin
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

package reuse

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/matrixorigin/matrixone/pkg/common/malloc"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
)

type mpoolBased[T any, P ReusableObject[T]] struct {
	pool  *mpool.MPool
	opts  *Options[T, P]
	c     *checker[T, P]
	inuse sync.Map // unsafe.Pointer -> malloc.Deallocator
}

func newMpoolBased[T any, P ReusableObject[T]](
	capacity int64,
	opts *Options[T, P]) Pool[T, P] {
	opts.adjust()
	c := newChecker[T, P](opts.enableChecker)
	var v T
	mp, err := mpool.NewMPool(fmt.Sprintf("reuse-%s", P(&v).TypeName()), opts.memCapacity, 0)
	if err != nil {
		panic(err)
	}
	return &mpoolBased[T, P]{
		pool: mp,
		opts: opts,
		c:    c,
	}
}

func (p *mpoolBased[T, P]) Alloc() P {
	var t T
	data, dec, err := p.pool.Alloc(int(unsafe.Sizeof(t)))
	if err != nil {
		panic(err)
	}
	ptr := unsafe.Pointer(unsafe.SliceData(data))
	p.inuse.Store(ptr, dec)
	v := (*T)(ptr)
	p.c.created(v)
	p.c.got(v)
	return v
}

func (p *mpoolBased[T, P]) Free(v P) {
	p.c.free(v)
	p.opts.release(v)
	ptr := unsafe.Pointer(v)
	value, ok := p.inuse.LoadAndDelete(ptr)
	if !ok {
		panic("bad pointer")
	}
	deallocator := value.(malloc.Deallocator)
	p.pool.Free(unsafe.Slice((*byte)(ptr), unsafe.Sizeof(*v)), deallocator)
}
