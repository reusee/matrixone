// Copyright 2023 Matrix Origin
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

package disttae

import "github.com/google/btree"

type btreeIter[T any] struct {
	tree     *btree.BTreeG[T]
	entries  []T
	i        int
	pivot    *T
	lessFunc btree.LessFunc[T]
}

const maxBTreeCachedEntries = 256

func newBTreeIter[T any](tree *btree.BTreeG[T], lessFunc btree.LessFunc[T]) *btreeIter[T] {
	return &btreeIter[T]{
		tree:     tree.Clone(),
		entries:  make([]T, 0, maxBTreeCachedEntries),
		lessFunc: lessFunc,
	}
}

func (b *btreeIter[T]) Next() bool {
	if b.i >= len(b.entries) {

		// load from pivot
		if b.pivot != nil {
			return b.Seek(*b.pivot)
		}

		// load from start
		b.i = 0
		b.entries = b.entries[:0]
		b.tree.Ascend(func(entry T) bool {
			b.entries = append(b.entries, entry)
			return len(b.entries) < maxBTreeCachedEntries
		})
		if len(b.entries) > 0 {
			lastEntry := b.entries[len(b.entries)-1]
			b.pivot = &lastEntry
		}

	} else {
		b.i++
	}

	return b.i < len(b.entries)
}

func (b *btreeIter[T]) Seek(pivot T) bool {
	b.i = 0
	b.entries = b.entries[:0]
	b.tree.AscendGreaterOrEqual(pivot, func(entry T) bool {
		if !b.lessFunc(entry, pivot) &&
			!b.lessFunc(pivot, entry) {
			// equal to pivot, skip
			return true
		}
		b.entries = append(b.entries, entry)
		return len(b.entries) < maxBTreeCachedEntries
	})
	if len(b.entries) > 0 {
		lastEntry := b.entries[len(b.entries)-1]
		b.pivot = &lastEntry
	}
	return b.i < len(b.entries)
}

func (b *btreeIter[T]) Entry() T {
	return b.entries[b.i]
}
