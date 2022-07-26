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
	Attrs any,
] struct {
	sync.Mutex
	Rows *btree.BTreeG[*Row[PrimaryKey, Attrs]]
}

type Row[
	PrimaryKey Ordered[PrimaryKey],
	Attrs any,
] struct {
	PrimaryKey PrimaryKey
	Values     *MVCC[Attrs]
}

type Ordered[To any] interface {
	Less(to To) bool
}

func NewTable[
	PrimaryKey Ordered[PrimaryKey],
	Attrs any,
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
	key PrimaryKey,
	attrs Attrs,
) error {
	t.Lock()
	row := t.getRow(key)
	t.Unlock()
	row.Values.Insert(tx, writeTime, attrs)
	return nil
}

func (t *Table[PrimaryKey, Attrs]) Update(
	tx *Transaction,
	writeTime Timestamp,
	key PrimaryKey,
	attrs Attrs,
) error {
	t.Lock()
	row := t.getRow(key)
	t.Unlock()
	row.Values.Update(tx, writeTime, attrs)
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
