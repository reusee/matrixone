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
	"unsafe"

	"golang.org/x/sys/unix"
)

type ReadOnlyAllocator struct {
	upstream        Allocator
	deallocatorPool *ClosureDeallocatorPool[readOnlyDeallocatorArgs]
}

type readOnlyDeallocatorArgs struct {
	info   MmapInfo
	frozen bool
}

func (r readOnlyDeallocatorArgs) As(trait Trait) bool {
	if ptr, ok := trait.(*Freeze); ok {
		*ptr = r.freeze
		return true
	}
	return false
}

func (r *readOnlyDeallocatorArgs) freeze() {
	r.frozen = true
	slice := unsafe.Slice(
		(*byte)(r.info.Addr),
		r.info.Length,
	)
	unix.Mprotect(slice, unix.PROT_READ)
}

type Freeze func()

func (*Freeze) IsTrait() {}

func NewReadOnlyAllocator(
	upstream Allocator,
) *ReadOnlyAllocator {
	return &ReadOnlyAllocator{
		upstream: upstream,

		deallocatorPool: NewClosureDeallocatorPool(
			func(hints Hints, args *readOnlyDeallocatorArgs) {
				if args.frozen {
					// unfreeze
					slice := unsafe.Slice(
						(*byte)(args.info.Addr),
						args.info.Length,
					)
					unix.Mprotect(slice, unix.PROT_READ|unix.PROT_WRITE)
				}
			},
		),
	}
}

var _ Allocator = new(ReadOnlyAllocator)

func (r *ReadOnlyAllocator) Allocate(size uint64, hint Hints) ([]byte, Deallocator, error) {
	bytes, dec, err := r.upstream.Allocate(size, hint)
	if err != nil {
		return nil, nil, err
	}

	var args readOnlyDeallocatorArgs
	if !dec.As(&args.info) {
		// not mmap allocated
		return bytes, dec, nil
	}

	return bytes, ChainDeallocator(
		dec,
		r.deallocatorPool.Get(args),
	), nil
}
