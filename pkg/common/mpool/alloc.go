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
	"unsafe"

	"github.com/matrixorigin/matrixone/pkg/common/malloc"
)

func alloc(allocator malloc.Allocator, sz, requiredSpaceWithoutHeader int, mp *MPool) ([]byte, malloc.Deallocator, error) {
	size := requiredSpaceWithoutHeader + kMemHdrSz
	slice, dec, err := allocator.Allocate(uint64(size), malloc.NoHints)
	if err != nil {
		return nil, nil, err
	}

	header := (*memHdr)(unsafe.Pointer(unsafe.SliceData(slice)))
	header.poolId = mp.id
	header.fixedPoolIdx = NumFixedPool
	header.allocSz = int32(sz)
	header.SetGuard()

	if mp.details != nil {
		mp.details.recordAlloc(int64(header.allocSz))
	}

	slice = slice[kMemHdrSz:][:requiredSpaceWithoutHeader:requiredSpaceWithoutHeader]

	return slice, dec, nil
}

type noopDeallocator struct{}

var _ malloc.Deallocator = noopDeallocator{}

func (n noopDeallocator) Deallocate(hints malloc.Hints) {}

func (n noopDeallocator) As(trait malloc.Trait) bool {
	return false
}

var NoopDeallocator noopDeallocator
