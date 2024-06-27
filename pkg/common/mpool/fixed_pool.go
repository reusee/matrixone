// Copyright 2021 - 2024 Matrix Origin
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
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
)

const (
	NumFixedPool = 5
)

// Pool emement size
var PoolElemSize = [NumFixedPool]int32{64, 128, 256, 512, 1024}

// pool for fixed elements.  Note that we preconfigure the pool size.
// We should consider implement some kind of growing logic.
type fixedPool struct {
	allocator malloc.Allocator
	m         sync.Mutex
	noLock    bool
	fpIdx     int8
	poolId    int64
	eleSz     int32
	// holds buffers allocated, it is not really used in alloc/free
	// but hold here for bookkeeping.
	buf             [][]byte
	bufDeallocators []malloc.Deallocator
	flist           unsafe.Pointer
}

// Initaialze a fixed pool
func (fp *fixedPool) initPool(allocator malloc.Allocator, tag string, poolid int64, idx int, noLock bool) {
	fp.allocator = allocator
	eleSz := PoolElemSize[idx]
	fp.poolId = poolid
	fp.fpIdx = int8(idx)
	fp.noLock = noLock
	fp.eleSz = eleSz
}

func (fp *fixedPool) closePool() {
	for _, dec := range fp.bufDeallocators {
		dec.Deallocate(malloc.NoHints)
	}
}

func (fp *fixedPool) nextPtr(ptr unsafe.Pointer) unsafe.Pointer {
	iptr := *(*unsafe.Pointer)(unsafe.Add(ptr, kMemHdrSz))
	return iptr
}

func (fp *fixedPool) setNextPtr(ptr unsafe.Pointer, next unsafe.Pointer) {
	iptr := (*unsafe.Pointer)(unsafe.Add(ptr, kMemHdrSz))
	*iptr = next
}

func (fp *fixedPool) alloc(sz int32) *memHdr {
	if !fp.noLock {
		fp.m.Lock()
		defer fp.m.Unlock()
	}

	if fp.flist == nil {
		size := kStripeSize * (fp.eleSz + kMemHdrSz)
		buf, dec, err := fp.allocator.Allocate(uint64(size), malloc.NoHints)
		if err != nil {
			panic(err)
		}
		ptr := unsafe.Pointer(unsafe.SliceData(buf))

		fp.buf = append(fp.buf, buf)
		fp.bufDeallocators = append(fp.bufDeallocators, dec)
		// return the first one
		ret := ptr
		header := (*memHdr)(ret)
		header.poolId = fp.poolId
		header.allocSz = sz
		header.fixedPoolIdx = fp.fpIdx
		header.SetGuard()

		// and thread the rest
		ptr = unsafe.Add(ret, fp.eleSz+kMemHdrSz)
		for i := 1; i < kStripeSize; i++ {
			header := (*memHdr)(ptr)
			header.poolId = fp.poolId
			header.allocSz = -1
			header.fixedPoolIdx = fp.fpIdx
			header.SetGuard()
			fp.setNextPtr(ptr, fp.flist)
			fp.flist = ptr
			ptr = unsafe.Add(ptr, fp.eleSz+kMemHdrSz)
		}

		return (*memHdr)(ret)

	} else {
		ret := fp.flist
		fp.flist = fp.nextPtr(fp.flist)
		header := (*memHdr)(ret)
		header.allocSz = sz
		// Zero slice.  Go requires slice to be zeroed.
		clear(unsafe.Slice((*byte)(unsafe.Add(ret, kMemHdrSz)), fp.eleSz))
		return header
	}
}

func (fp *fixedPool) free(hdr *memHdr) {
	if hdr.poolId != fp.poolId || hdr.fixedPoolIdx != fp.fpIdx ||
		hdr.allocSz < 0 || hdr.allocSz > fp.eleSz ||
		!hdr.CheckGuard() {
		panic(moerr.NewInternalErrorNoCtx("mpool fixed pool hdr corruption.   Possible double free"))
	}

	if !fp.noLock {
		fp.m.Lock()
		defer fp.m.Unlock()
	}
	ptr := unsafe.Pointer(hdr)
	fp.setNextPtr(ptr, fp.flist)
	fp.flist = ptr
}
