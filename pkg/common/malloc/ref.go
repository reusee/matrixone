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
	"sync/atomic"
)

type Ref[T any] struct {
	id     uint64
	Value  T
	holder *RefHolder[T]
	role   HolderRole
}

type RefHolder[T any] struct {
	mu         sync.Mutex
	owns       map[uint64]bool
	borrowTo   map[uint64][]*RefHolder[T]
	borrowFrom map[uint64]*RefHolder[T]
}

type HolderRole uint8

const (
	Owner HolderRole = iota + 1
	Borrower
)

var nextID atomic.Uint64

func (r *RefHolder[T]) Own(value T) Ref[T] {
	id := nextID.Add(1)
	r.mu.Lock()
	r.owns[id] = true
	r.mu.Unlock()
	return Ref[T]{
		id:     id,
		Value:  value,
		holder: r,
		role:   Owner,
	}
}

func (r *RefHolder[T]) Borrow(ref Ref[T], to *RefHolder[T]) Ref[T] {
	r.mu.Lock()
	r.borrowTo[ref.id] = append(
		r.borrowTo[ref.id],
		to,
	)
	r.mu.Unlock()

	to.mu.Lock()
	if _, ok := to.borrowFrom[ref.id]; ok {
		panic("already borrowed")
	}
	to.borrowFrom[ref.id] = r
	to.mu.Unlock()

	return Ref[T]{
		id:     ref.id,
		Value:  ref.Value,
		holder: to,
		role:   Borrower,
	}
}

//TODO RefHolder end
//TODO Ref end
//TODO Ref move
