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

package disttae

import (
	"context"

	"github.com/matrixorigin/matrixone/pkg/pb/api"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/txn/storage/txn/memtable"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

type mvcc struct {
	table *memtable.Table[PrimaryKey, *Row, *Row]
}

type PrimaryKey struct {
	//TODO
}

var _ memtable.Ordered[PrimaryKey] = PrimaryKey{}

func (p PrimaryKey) Less(than PrimaryKey) bool {
	//TODO
	return false
}

type Row struct {
	key PrimaryKey
	//TODO
}

var _ memtable.Row[PrimaryKey, *Row] = new(Row)

func (r *Row) Key() PrimaryKey {
	return r.key
}

func (r *Row) Value() *Row {
	return r
}

func (r *Row) Indexes() []memtable.Tuple {
	return nil
}

var _ MVCC = new(mvcc)

// BlockList implements MVCC
func (*mvcc) BlockList(ctx context.Context, ts timestamp.Timestamp, entries [][]Entry) []BlockMeta {
	panic("unimplemented")
}

// CheckPoint implements MVCC
func (*mvcc) CheckPoint(ts timestamp.Timestamp) error {
	panic("unimplemented")
}

// Delete implements MVCC
func (*mvcc) Delete(ctx context.Context, bat *api.Batch) error {
	panic("unimplemented")
}

// Insert implements MVCC
func (*mvcc) Insert(ctx context.Context, bat *api.Batch) error {
	panic("unimplemented")
}

// NewReader implements MVCC
func (*mvcc) NewReader(ctx context.Context, readerNumber int, expr *plan.Expr, ts timestamp.Timestamp, entries [][]Entry) ([]engine.Reader, error) {
	panic("unimplemented")
}
