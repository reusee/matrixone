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
	"sync/atomic"
	"unsafe"

	"github.com/matrixorigin/matrixone/pkg/util/missingfreeguard"
)

func alloc(sz, requiredSpaceWithoutHeader int, mp *MPool) []byte {
	bs := make([]byte, requiredSpaceWithoutHeader+kMemHdrSz)
	hdr := unsafe.Pointer(&bs[0])
	pHdr := (*memHdr)(hdr)
	pHdr.poolId = mp.id
	pHdr.fixedPoolIdx = NumFixedPool
	pHdr.allocSz = int32(sz)
	pHdr.SetGuard()

	id := nextID.Add(1)
	pHdr.guardID = id
	guard := guardManager.NewGuard(&bs[0], int64(sz))
	guardsMap.Store(id, guard)

	if mp.details != nil {
		mp.details.recordAlloc(int64(pHdr.allocSz))
	}
	return pHdr.ToSlice(sz, requiredSpaceWithoutHeader)
}

var guardManager = missingfreeguard.NewManager("/missing-free-mpool/")

var guardsMap sync.Map

var nextID atomic.Int64

func freeGuard(id int64) {
	v, deleted := guardsMap.LoadAndDelete(id)
	if deleted {
		v.(*missingfreeguard.Guard).Free()
	}
}
