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
	"slices"
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

func NewRefHolder[T any]() *RefHolder[T] {
	return &RefHolder[T]{
		owns:       make(map[uint64]bool),
		borrowTo:   make(map[uint64][]*RefHolder[T]),
		borrowFrom: make(map[uint64]*RefHolder[T]),
	}
}

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

func (r *RefHolder[T]) borrow(ref Ref[T], to *RefHolder[T]) Ref[T] {
	if to == r {
		panic("borrow to owner")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	to.mu.Lock()
	defer to.mu.Unlock()

	r.borrowTo[ref.id] = append(
		r.borrowTo[ref.id],
		to,
	)
	if _, ok := to.borrowFrom[ref.id]; ok {
		panic("already borrowed")
	}
	to.borrowFrom[ref.id] = r

	return Ref[T]{
		id:     ref.id,
		Value:  ref.Value,
		holder: to,
		role:   Borrower,
	}
}

func (r Ref[T]) Borrow(holder *RefHolder[T]) Ref[T] {
	if r.role != Owner {
		panic("cannot borrow")
	}
	return r.holder.borrow(r, holder)
}

func (r *Ref[T]) End() {
	if r.id == 0 ||
		r.holder == nil ||
		r.role == 0 {
		panic("null Ref")
	}

	defer func() {
		*r = Ref[T]{}
	}()

	switch r.role {

	case Owner:
		r.holder.mu.Lock()
		defer r.holder.mu.Unlock()
		delete(r.holder.owns, r.id)
		borrows := r.holder.borrowTo[r.id]
		delete(r.holder.borrowTo, r.id)
		if len(borrows) > 0 {
			panic("still being borrowed")
		}

	case Borrower:
		r.holder.mu.Lock()
		defer r.holder.mu.Unlock()
		owner := r.holder.borrowFrom[r.id]
		delete(r.holder.borrowFrom, r.id)
		if owner == nil {
			panic("owner not found")
		}
		owner.mu.Lock()
		defer owner.mu.Unlock()
		owner.borrowTo[r.id] = slices.DeleteFunc(
			owner.borrowTo[r.id],
			func(h *RefHolder[T]) bool {
				return h == r.holder
			},
		)

	default:
		panic("invalid role")
	}
}

//TODO RefHolder end
//TODO Ref move
