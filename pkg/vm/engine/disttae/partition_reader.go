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
	"fmt"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/txn/storage/memorystorage/memtable"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

type PartitionReader struct {
	end           bool
	typsMap       map[string]types.Type
	firstCalled   bool
	readTimestamp timestamp.Timestamp
	readTime      memtable.Time
	tx            *memtable.Transaction
	index         memtable.Tuple
	inserts       []*batch.Batch
	deletes       map[types.Rowid]uint8
	skipBlocks    map[uint64]uint8
	iter          *PartitionIter
	partition     *Partition
}

var _ engine.Reader = new(PartitionReader)

func (p *PartitionReader) Close() error {
	return nil
}

func (p *PartitionReader) Read(ctx context.Context, colNames []string, expr *plan.Expr, mp *mpool.MPool) (*batch.Batch, error) {
	if p == nil {
		return nil, nil
	}
	if p.end {
		return nil, nil
	}
	if len(p.inserts) > 0 {
		bat := p.inserts[0].GetSubBatch(colNames)
		p.inserts = p.inserts[1:]
		b := batch.NewWithSize(len(colNames))
		b.SetAttributes(colNames)
		for i, name := range colNames {
			b.Vecs[i] = vector.New(p.typsMap[name])
		}
		if _, err := b.Append(ctx, mp, bat); err != nil {
			return nil, err
		}
		return b, nil
	}
	b := batch.NewWithSize(len(colNames))
	b.SetAttributes(colNames)
	for i, name := range colNames {
		b.Vecs[i] = vector.New(p.typsMap[name])
	}
	rows := 0
	if len(p.index) > 0 {
		itr := p.partition.NewIter(p.readTimestamp, p.index, p.index)
		for ok := itr.First(); ok; ok = itr.Next() {
			rowID, dataValue := itr.Item()
			if _, ok := p.deletes[rowID]; ok {
				continue
			}
			if p.skipBlocks != nil {
				if _, ok := p.skipBlocks[rowIDToBlockID(rowID)]; ok {
					continue
				}
			}
			if dataValue.op == opDelete {
				continue
			}
			for i, name := range b.Attrs {
				if name == catalog.Row_ID {
					if err := b.Vecs[i].Append(rowID, false, mp); err != nil {
						return nil, err
					}
					continue
				}
				value, ok := dataValue.value[name]
				if !ok {
					panic(fmt.Sprintf("invalid column name: %v", name))
				}
				if err := value.AppendVector(b.Vecs[i], mp); err != nil {
					return nil, err
				}
			}
			rows++
		}
		if rows > 0 {
			b.SetZs(rows, mp)
		}
		itr.Close()
		p.end = true
		if rows == 0 {
			return nil, nil
		}
		return b, nil
	}

	fn := p.iter.Next
	if !p.firstCalled {
		fn = p.iter.First
		p.firstCalled = true
	}

	maxRows := 8192 // i think 8192 is better than 4096
	for ok := fn(); ok; ok = p.iter.Next() {
		rowID, dataValue := p.iter.Item()

		if _, ok := p.deletes[rowID]; ok {
			continue
		}

		if dataValue.op == opDelete {
			continue
		}

		if p.skipBlocks != nil {
			if _, ok := p.skipBlocks[rowIDToBlockID(rowID)]; ok {
				continue
			}
		}

		for i, name := range b.Attrs {
			if name == catalog.Row_ID {
				if err := b.Vecs[i].Append(rowID, false, mp); err != nil {
					return nil, err
				}
				continue
			}
			value, ok := dataValue.value[name]
			if !ok {
				panic(fmt.Sprintf("invalid column name: %v", name))
			}
			if err := value.AppendVector(b.Vecs[i], mp); err != nil {
				return nil, err
			}
		}

		rows++
		if rows == maxRows {
			break
		}
	}

	if rows > 0 {
		b.SetZs(rows, mp)
	}
	if rows == 0 {
		return nil, nil
	}

	return b, nil
}
