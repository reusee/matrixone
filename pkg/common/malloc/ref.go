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
	_      noCopy
	id     uint64
	Value  T
	holder *RefHolder[T]
	role   HolderRole
}

type RefHolder[T any] struct {
	mu      sync.Mutex
	owns    map[uint64]bool
	lends   map[uint64][]*RefHolder[T]
	borrows map[uint64]*RefHolder[T]
}

type HolderRole uint8

const (
	Owner HolderRole = iota + 1
	Borrower
)

func NewRefHolder[T any]() *RefHolder[T] {
	return &RefHolder[T]{
		owns:    make(map[uint64]bool),
		lends:   make(map[uint64][]*RefHolder[T]),
		borrows: make(map[uint64]*RefHolder[T]),
	}
}

var nextID atomic.Uint64

func (r *RefHolder[T]) Own(value T) *Ref[T] {
	id := nextID.Add(1)
	r.mu.Lock()
	r.owns[id] = true
	r.mu.Unlock()
	return &Ref[T]{
		id:     id,
		Value:  value,
		holder: r,
		role:   Owner,
	}
}

func (r *RefHolder[T]) borrow(ref *Ref[T], to *RefHolder[T]) *Ref[T] {
	if to == r {
		panic("borrow to owner")
	}

	r.mu.Lock()
	r.lends[ref.id] = append(
		r.lends[ref.id],
		to,
	)
	r.mu.Unlock()

	to.mu.Lock()
	if _, ok := to.borrows[ref.id]; ok {
		to.mu.Unlock()
		panic("already lent")
	}
	to.borrows[ref.id] = r
	to.mu.Unlock()

	return &Ref[T]{
		id:     ref.id,
		Value:  ref.Value,
		holder: to,
		role:   Borrower,
	}
}

func (r *Ref[T]) Borrow(holder *RefHolder[T]) *Ref[T] {
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
		delete(r.holder.owns, r.id)
		borrows := r.holder.lends[r.id]
		delete(r.holder.lends, r.id)
		r.holder.mu.Unlock()
		if len(borrows) > 0 {
			panic("still being lent")
		}

	case Borrower:
		r.holder.mu.Lock()
		owner := r.holder.borrows[r.id]
		delete(r.holder.borrows, r.id)
		r.holder.mu.Unlock()
		if owner == nil {
			panic("owner not found")
		}

		owner.mu.Lock()
		owner.lends[r.id] = slices.DeleteFunc(
			owner.lends[r.id],
			func(h *RefHolder[T]) bool {
				return h == r.holder
			},
		)
		owner.mu.Unlock()

	default:
		panic("invalid role")
	}
}

func (r *RefHolder[T]) move(ref *Ref[T], to *RefHolder[T]) {
	if ref.holder != r {
		panic("not holder")
	}
	if r == to {
		panic("same holder")
	}

	ref.holder = to

	// delete from r
	var borrowers []*RefHolder[T]
	var owner *RefHolder[T]
	r.mu.Lock()
	switch ref.role {

	case Owner:
		delete(r.owns, ref.id)
		borrowers = r.lends[ref.id]
		delete(r.lends, ref.id)

	case Borrower:
		var ok bool
		owner, ok = r.borrows[ref.id]
		if !ok {
			r.mu.Unlock()
			panic("owner not found")
		}
		delete(r.borrows, ref.id)

	default:
		r.mu.Unlock()
		panic("invalid role")
	}
	r.mu.Unlock()

	// update to
	to.mu.Lock()
	switch ref.role {

	case Owner:
		to.owns[ref.id] = true
		for _, borrower := range borrowers {
			if borrower == to {
				to.mu.Unlock()
				panic("cannot move ownership to borrower")
			}
			to.lends[ref.id] = append(
				to.lends[ref.id],
				borrower,
			)
		}

	case Borrower:
		to.borrows[ref.id] = owner

	default:
		to.mu.Unlock()
		panic("invalid role")
	}
	to.mu.Unlock()

	// update borrowers
	for _, borrower := range borrowers {
		borrower.mu.Lock()
		borrower.borrows[ref.id] = to
		borrower.mu.Unlock()
	}

	// update owner
	if owner != nil {
		owner.mu.Lock()
		for i, borrower := range owner.lends[ref.id] {
			if borrower == r {
				owner.lends[ref.id][i] = to
			}
		}
		owner.mu.Unlock()
	}

}

func (r *Ref[T]) Move(to *RefHolder[T]) {
	r.holder.move(r, to)
}

func (r *RefHolder[T]) End() {
	r.mu.Lock()

	// check owns
	if len(r.owns) > 0 {
		r.mu.Unlock()
		panic("end with live owned references")
	}

	// check lends
	if len(r.lends) > 0 {
		r.mu.Unlock()
		panic("end with live lent references")
	}

	// borrows
	infos := make([]borrowInfo[T], 0, len(r.borrows))
	for id, owner := range r.borrows {
		infos = append(infos, borrowInfo[T]{
			id:    id,
			owner: owner,
		})
	}

	r.mu.Unlock()

	// clear
	*r = RefHolder[T]{}

	// update owner lends
	for _, info := range infos {
		info.owner.mu.Lock()
		info.owner.lends[info.id] = slices.DeleteFunc(
			info.owner.lends[info.id],
			func(h *RefHolder[T]) bool {
				return h == r
			},
		)
		info.owner.mu.Unlock()
	}

}

type borrowInfo[T any] struct {
	id    uint64
	owner *RefHolder[T]
}
