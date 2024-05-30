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
	"fmt"
	"math/bits"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	MB = 1 << 20
	GB = 1 << 30

	// There are 14 classes under this threshold,
	// so the default sharded allocator will consume less than 14MB per CPU core.
	// On a 256 core machine, it will be 3584MB, not too high.
	// We may tune this if required.
	smallClassCap = 1 * MB

	// A large class size will not consume more memory unless we actually allocate on that class
	// Classes with size larger than smallClassCap will always buffering MADV_DONTNEED-advised objects.
	// MADV_DONTNEED-advised objects consume no memory.
	maxClassSize = 4 * GB

	maxBuffer1Cap = 256
)

type ClassAllocator struct {
	classes []Class
}

type Class struct {
	size      uint64
	allocator *fixedSizeMmapAllocator
}

func NewClassAllocator(
	checkFraction uint32,
) *ClassAllocator {
	ret := &ClassAllocator{}

	// init classes
	for size := uint64(1); size <= maxClassSize; size *= 2 {
		ret.classes = append(ret.classes, Class{
			size:      size,
			allocator: newFixedSizedAllocator(size, checkFraction),
		})
	}

	return ret
}

var _ Allocator = new(ClassAllocator)

func (c *ClassAllocator) Allocate(size uint64) (unsafe.Pointer, Deallocator) {
	if size == 0 {
		panic("invalid size: 0")
	}
	// class size factor is 2, so we can calculate the class index
	var i int
	if bits.OnesCount64(size) > 1 {
		// round to next bucket
		i = bits.Len64(size)
	} else {
		// power of two
		i = bits.Len64(size) - 1
	}
	if i >= len(c.classes) {
		panic(fmt.Sprintf("cannot allocate %v bytes: too large", size))
	}
	return c.classes[i].allocator.Allocate(size)
}

type fixedSizeMmapAllocator struct {
	size          uint64
	checkFraction uint32
	// buffer1 buffers objects
	buffer1 chan unsafe.Pointer
	bump    uint64
	mem     atomic.Pointer[mappedMemory]
}

type mappedMemory struct {
	mem  []byte
	next atomic.Uint64
}

func newFixedSizedAllocator(
	size uint64,
	checkFraction uint32,
) *fixedSizeMmapAllocator {

	// if size is larger than smallClassCap, num1 will be zero, buffer1 will be empty
	num1 := smallClassCap / size
	if num1 > maxBuffer1Cap {
		// don't buffer too much, since chans with larger buffer consume more memory
		num1 = maxBuffer1Cap
	}

	bump := size
	if bump < 4096 {
		// align to page to allow madvise
		bump = 4096
	}
	ret := &fixedSizeMmapAllocator{
		size:          size,
		checkFraction: checkFraction,
		buffer1:       make(chan unsafe.Pointer, num1),
		bump:          bump,
	}

	mem, err := unix.Mmap(
		-1, 0,
		maxClassSize,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_PRIVATE|unix.MAP_ANONYMOUS,
	)
	if err != nil {
		panic(err)
	}
	ret.mem.Store(&mappedMemory{
		mem: mem,
	})

	return ret
}

var _ Allocator = new(fixedSizeMmapAllocator)

func (f *fixedSizeMmapAllocator) Allocate(_ uint64) (ptr unsafe.Pointer, dec Deallocator) {
	if f.checkFraction > 0 {
		defer func() {
			if fastrand()%f.checkFraction == 0 {
				dec = newCheckedDeallocator(dec)
			}
		}()
	}

	select {

	case ptr := <-f.buffer1:
		// from buffer1
		clear(unsafe.Slice((*byte)(ptr), f.size))
		return ptr, f

	default:
		for {

			// bump
			mapped := f.mem.Load()
			offset := mapped.next.Add(f.bump)
			if offset+f.size < uint64(len(mapped.mem)) {
				return unsafe.Pointer(unsafe.SliceData(mapped.mem[offset:])), f
			}

			// no more space
			mem, err := unix.Mmap(
				-1, 0,
				maxClassSize,
				unix.PROT_READ|unix.PROT_WRITE,
				unix.MAP_PRIVATE|unix.MAP_ANONYMOUS,
			)
			if err != nil {
				panic(err)
			}
			swapped := f.mem.CompareAndSwap(mapped, &mappedMemory{
				mem: mem,
			})
			if !swapped {
				if err := unix.Munmap(mem); err != nil {
					panic(err)
				}
			}
		}

	}
}

var _ Deallocator = new(fixedSizeMmapAllocator)

func (f *fixedSizeMmapAllocator) Deallocate(ptr unsafe.Pointer) {

	if f.checkFraction > 0 &&
		fastrand()%f.checkFraction == 0 {
		// do not reuse to detect use-after-free
		if err := unix.Munmap(
			unsafe.Slice((*byte)(ptr), f.bump),
		); err != nil {
			panic(err)
		}
		return
	}

	select {

	case f.buffer1 <- ptr:
		// buffer in buffer1

	default:
		f.freeMem(ptr)

	}

}
