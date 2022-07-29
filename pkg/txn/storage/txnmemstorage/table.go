// Copyright 2022 Matrix Origin
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

package memstorage

import (
	"sync"

	"github.com/google/btree"
)

type Table[
	PrimaryKey Ordered[PrimaryKey],
	Attrs Attributes[PrimaryKey],
] struct {
	sync.Mutex
	Rows *btree.BTreeG[*Row[PrimaryKey, Attrs]]
}

type Attributes[PrimaryKey Ordered[PrimaryKey]] interface {
	PrimaryKey() PrimaryKey
}

type Row[
	PrimaryKey Ordered[PrimaryKey],
	Attrs Attributes[PrimaryKey],
] struct {
	PrimaryKey PrimaryKey
	Values     *MVCC[Attrs]
}

type Ordered[To any] interface {
	Less(to To) bool
}

func NewTable[
	PrimaryKey Ordered[PrimaryKey],
	Attrs Attributes[PrimaryKey],
]() *Table[PrimaryKey, Attrs] {
	return &Table[PrimaryKey, Attrs]{
		Rows: btree.NewG(2, func(a, b *Row[PrimaryKey, Attrs]) bool {
			return a.PrimaryKey.Less(b.PrimaryKey)
		}),
	}
}

func (t *Table[PrimaryKey, Attrs]) Insert(
	tx *Transaction,
	writeTime Timestamp,
	attrs Attrs,
) error {
	t.Lock()
	key := attrs.PrimaryKey()
	row := t.getRow(key)
	t.Unlock()
	row.Values.Insert(tx, writeTime, attrs)
	//TODO this is wrong
	// writeTime's logical time should be the statement number
	// but currently the engine does not expose statement numbers
	// for now, just tick on every write operation
	tx.Tick()
	return nil
}

func (t *Table[PrimaryKey, Attrs]) Update(
	tx *Transaction,
	writeTime Timestamp,
	attrs Attrs,
) error {
	t.Lock()
	key := attrs.PrimaryKey()
	row := t.getRow(key)
	t.Unlock()
	row.Values.Update(tx, writeTime, attrs)
	tx.Tick()
	return nil
}

func (t *Table[PrimaryKey, Attrs]) Delete(
	tx *Transaction,
	writeTime Timestamp,
	key PrimaryKey,
) error {
	t.Lock()
	row := t.getRow(key)
	t.Unlock()
	row.Values.Delete(tx, writeTime)
	tx.Tick()
	return nil
}

func (t *Table[PrimaryKey, Attrs]) Get(
	tx *Transaction,
	readTime Timestamp,
	key PrimaryKey,
) (
	attrs Attrs,
	err error,
) {
	t.Lock()
	row := t.getRow(key)
	t.Unlock()
	attrs = *row.Values.Read(tx, readTime)
	return
}

func (t *Table[PrimaryKey, Attrs]) getRow(key PrimaryKey) *Row[PrimaryKey, Attrs] {
	pivot := &Row[PrimaryKey, Attrs]{
		PrimaryKey: key,
	}
	row, ok := t.Rows.Get(pivot)
	if !ok {
		row = pivot
		row.Values = new(MVCC[Attrs])
		t.Rows.ReplaceOrInsert(row)
	}
	return row
}
