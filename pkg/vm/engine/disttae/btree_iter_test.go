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

import (
	"math/rand"
	"testing"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func TestBTreeIter(t *testing.T) {
	less := func(a, b int) bool {
		return a < b
	}
	tree := btree.NewG(2, less)
	for _, i := range rand.Perm(1024) {
		tree.ReplaceOrInsert(i)
	}
	iter := newBTreeIter(tree, less)
	last := -1
	for iter.Next() {
		entry := iter.Entry()
		assert.Equal(t, last+1, entry)
		last = entry
	}
	assert.Equal(t, 1023, last)

	for i := 0; i < 1024; i++ {
		ok := iter.Seek(i)
		assert.True(t, ok)
		entry := iter.Entry()
		assert.Equal(t, i, entry)
	}

}

func BenchmarkBTreeIter(b *testing.B) {
	less := func(a, b int) bool {
		return a < b
	}
	tree := btree.NewG(2, less)
	for _, i := range rand.Perm(1024) {
		tree.ReplaceOrInsert(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := newBTreeIter(tree, less)
		for iter.Next() {
		}
	}
}
