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
	m      sync.Mutex
	noLock bool
	fpIdx  int8
	poolId int64
	eleSz  int32
	// holds buffers allocated, it is not really used in alloc/free
	// but hold here for bookkeeping.
	buf   [][]byte
	flist unsafe.Pointer
}

// Initaialze a fixed pool
func (fp *fixedPool) initPool(tag string, poolid int64, idx int, noLock bool) {
	eleSz := PoolElemSize[idx]
	fp.poolId = poolid
	fp.fpIdx = int8(idx)
	fp.noLock = noLock
	fp.eleSz = eleSz
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
		buf := make([]byte, kStripeSize*(fp.eleSz+kMemHdrSz))
		fp.buf = append(fp.buf, buf)
		// return the first one
		ret := (unsafe.Pointer)(&buf[0])
		pHdr := (*memHdr)(ret)
		pHdr.poolId = fp.poolId
		pHdr.allocSz = sz
		pHdr.fixedPoolIdx = fp.fpIdx
		pHdr.SetGuard()

		ptr := unsafe.Add(ret, fp.eleSz+kMemHdrSz)
		// and thread the rest
		for i := 1; i < kStripeSize; i++ {
			pHdr := (*memHdr)(ptr)
			pHdr.poolId = fp.poolId
			pHdr.allocSz = -1
			pHdr.fixedPoolIdx = fp.fpIdx
			pHdr.SetGuard()
			fp.setNextPtr(ptr, fp.flist)
			fp.flist = ptr
			ptr = unsafe.Add(ptr, fp.eleSz+kMemHdrSz)
		}
		return (*memHdr)(ret)
	} else {
		ret := fp.flist
		fp.flist = fp.nextPtr(fp.flist)
		pHdr := (*memHdr)(ret)
		pHdr.allocSz = sz
		// Zero slice.  Go requires slice to be zeroed.
		bs := unsafe.Slice((*byte)(unsafe.Add(ret, kMemHdrSz)), fp.eleSz)
		// the compiler will optimize this loop to memclr
		for i := range bs {
			bs[i] = 0
		}
		return pHdr
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
