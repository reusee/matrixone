// Copyright 2024 Matrix Origin
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

package malloc

import (
	"sync"
	"unsafe"
	"weak"

	"golang.org/x/sys/cpu"
)

type ManagedAllocator[U Allocator] struct {
	upstream U
	inUse    [256]managedAllocatorShard
}

type managedAllocatorShard struct {
	sync.Mutex
	items map[unsafe.Pointer]weak.Pointer[Deallocator]
	_     cpu.CacheLinePad
}

func NewManagedAllocator[U Allocator](
	upstream U,
) *ManagedAllocator[U] {
	ret := &ManagedAllocator[U]{
		upstream: upstream,
	}
	for i := range len(ret.inUse) {
		ret.inUse[i].items = make(map[unsafe.Pointer]weak.Pointer[Deallocator])
	}
	return ret
}

func (m *ManagedAllocator[U]) Allocate(size uint64, hints Hints) ([]byte, error) {
	slice, dec, err := m.upstream.Allocate(size, hints)
	if err != nil {
		return nil, err
	}
	ptr := unsafe.Pointer(unsafe.SliceData(slice))
	shard := &m.inUse[hashPointer(uintptr(ptr))]
	shard.allocate(ptr, dec)
	return slice, nil
}

func (m *ManagedAllocator[U]) Deallocate(slice []byte, hints Hints) {
	ptr := unsafe.Pointer(unsafe.SliceData(slice))
	shard := &m.inUse[hashPointer(uintptr(ptr))]
	shard.deallocate(ptr, hints)
}

func (m *managedAllocatorShard) allocate(ptr unsafe.Pointer, deallocator Deallocator) {
	m.Lock()
	defer m.Unlock()
	m.items[ptr] = weak.Make(&deallocator)
}

func (m *managedAllocatorShard) deallocate(ptr unsafe.Pointer, hints Hints) {
	m.Lock()
	defer m.Unlock()
	deallocator, ok := m.items[ptr]
	if !ok {
		panic("bad pointer")
	} else {
		(*deallocator.Value()).Deallocate(hints)
		delete(m.items, ptr)
	}
}

func hashPointer(ptr uintptr) uint8 {
	ret := uint8(ptr)
	ret ^= uint8(ptr >> 8)
	ret ^= uint8(ptr >> 16)
	ret ^= uint8(ptr >> 24)
	ret ^= uint8(ptr >> 32)
	ret ^= uint8(ptr >> 40)
	ret ^= uint8(ptr >> 48)
	ret ^= uint8(ptr >> 56)
	return ret
}
