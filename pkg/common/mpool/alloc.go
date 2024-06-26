// Copyright 2021 - 2022 Matrix Origin
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

package mpool

import (
	"sync"
	"unsafe"

	"github.com/matrixorigin/matrixone/pkg/common/malloc"
)

var ptrToDeallocator sync.Map

func alloc(allocator malloc.Allocator, sz, requiredSpaceWithoutHeader int, mp *MPool) []byte {
	size := requiredSpaceWithoutHeader + kMemHdrSz
	ptr, dec, err := allocator.Allocate(uint64(size), malloc.NoHints)
	if err != nil {
		panic(err)
	}

	header := (*memHdr)(ptr)
	header.poolId = mp.id
	header.fixedPoolIdx = NumFixedPool
	header.allocSz = int32(sz)
	header.SetGuard()

	if mp.details != nil {
		mp.details.recordAlloc(int64(header.allocSz))
	}

	slice := unsafe.Slice(
		(*byte)(unsafe.Add(ptr, kMemHdrSz)),
		requiredSpaceWithoutHeader,
	)[:sz]

	ptrToDeallocator.Store(ptr, dec)

	return slice
}

func free(ptr unsafe.Pointer) {
	v, ok := ptrToDeallocator.Load(ptr)
	if !ok {
		panic("bad pointer")
	}
	v.(malloc.Deallocator).Deallocate(ptr, malloc.NoHints)
}
