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

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/pb/api"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/txn/storage/memorystorage/memorytable"
	"github.com/matrixorigin/matrixone/pkg/txn/storage/memorystorage/memtable"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/tidwall/btree"
)

func NewPartition(
	columnsIndexDefs []ColumnsIndexDef,
) *Partition {
	return &Partition{
		index:            btree.NewBTreeG(comparePartitionIndexEntry),
		columnsIndexDefs: columnsIndexDefs,
	}
}

type DataValue struct {
	op    uint8
	value map[string]memtable.Nullable
}

const (
	opInsert = iota + 1
	opDelete
)

var _ MVCC = new(Partition)

func (p *Partition) BlockList(
	ctx context.Context,
	ts timestamp.Timestamp,
	blocks []BlockMeta,
	entries []Entry,
) (
	[]BlockMeta,
	map[uint64][]int,
) {
	blks := make([]BlockMeta, 0, len(blocks))
	deletes := make(map[uint64][]int)
	if len(blocks) == 0 {
		return blks, deletes
	}
	ids := make([]uint64, len(blocks))
	for i := range blocks {
		// if cn can see a appendable block, this block must contain all updates
		// in cache, no need to do merge read, BlockRead will filter out
		// invisible and deleted rows with respect to the timestamp
		if !blocks[i].Info.EntryState {
			ids[i] = blocks[i].Info.BlockID
		}
	}
	p.IterDeletedRowIDs(ctx, ids, ts, func(rowID types.Rowid) bool {
		id, offset := catalog.DecodeRowid(types.Rowid(rowID))
		deletes[id] = append(deletes[id], int(offset))
		return true
	})
	for _, entry := range entries {
		if entry.typ == DELETE {
			vs := vector.MustTCols[types.Rowid](entry.bat.GetVector(0))
			for _, v := range vs {
				id, offset := catalog.DecodeRowid(v)
				deletes[id] = append(deletes[id], int(offset))
			}
		}
	}
	for i := range blocks {
		if _, ok := deletes[blocks[i].Info.BlockID]; !ok {
			blks = append(blks, blocks[i])
		}
	}
	return blks, deletes
}

func (*Partition) CheckPoint(ctx context.Context, ts timestamp.Timestamp) error {
	panic("unimplemented")
}

func (p *Partition) Delete(ctx context.Context, b *api.Batch) error {
	bat, err := batch.ProtoBatchToBatch(b)
	if err != nil {
		return err
	}
	p.addDeleteBatch(bat)
	return nil
}

func (p *Partition) Insert(ctx context.Context, primaryKeyIndex int,
	b *api.Batch, needCheck bool) error {
	bat, err := batch.ProtoBatchToBatch(b)
	if err != nil {
		return err
	}
	p.addInsertBatch(bat)
	return nil
}

func rowIDToBlockID(rowID types.Rowid) uint64 {
	id, _ := catalog.DecodeRowid(rowID)
	return id
}

func (p *Partition) NewReader(
	ctx context.Context,
	readerNumber int,
	index memtable.Tuple,
	defs []engine.TableDef,
	tableDef *plan.TableDef,
	skipBlocks map[uint64]uint8,
	blks []ModifyBlockMeta,
	ts timestamp.Timestamp,
	fs fileservice.FileService,
	entries []Entry,
) ([]engine.Reader, error) {

	t := memtable.Time{
		Timestamp: ts,
	}
	tx := memtable.NewTransaction(
		newMemTableTransactionID(),
		t,
		memtable.SnapshotIsolation,
	)

	inserts := make([]*batch.Batch, 0, len(entries))
	deletes := make(map[types.Rowid]uint8)
	for _, entry := range entries {
		if entry.typ == INSERT {
			inserts = append(inserts, entry.bat)
		} else {
			if entry.bat.GetVector(0).GetType().Oid == types.T_Rowid {
				vs := vector.MustTCols[types.Rowid](entry.bat.GetVector(0))
				for _, v := range vs {
					deletes[v] = 0
				}
			}
		}
	}

	readers := make([]engine.Reader, readerNumber)

	mp := make(map[string]types.Type)
	mp[catalog.Row_ID] = types.New(types.T_Rowid, 0, 0, 0)
	for _, def := range defs {
		attr, ok := def.(*engine.AttributeDef)
		if !ok {
			continue
		}
		mp[attr.Attr.Name] = attr.Attr.Type
	}

	readers[0] = &PartitionReader{
		typsMap:       mp,
		readTime:      t,
		readTimestamp: ts,
		tx:            tx,
		index:         index,
		inserts:       inserts,
		deletes:       deletes,
		skipBlocks:    skipBlocks,
		iter:          p.NewIter(ts, index, index),
		partition:     p,
	}
	if readerNumber == 1 {
		for i := range blks {
			readers = append(readers, &blockMergeReader{
				fs:       fs,
				ts:       ts,
				ctx:      ctx,
				tableDef: tableDef,
				sels:     make([]int64, 0, 1024),
				blks:     []ModifyBlockMeta{blks[i]},
			})
		}
		return []engine.Reader{&mergeReader{readers}}, nil
	}
	if len(blks) < readerNumber-1 {
		for i := range blks {
			readers[i+1] = &blockMergeReader{
				fs:       fs,
				ts:       ts,
				ctx:      ctx,
				tableDef: tableDef,
				sels:     make([]int64, 0, 1024),
				blks:     []ModifyBlockMeta{blks[i]},
			}
		}
		for j := len(blks) + 1; j < readerNumber; j++ {
			readers[j] = &emptyReader{}
		}
		return readers, nil
	}
	step := len(blks) / (readerNumber - 1)
	if step < 1 {
		step = 1
	}
	for i := 1; i < readerNumber; i++ {
		if i == readerNumber-1 {
			readers[i] = &blockMergeReader{
				fs:       fs,
				ts:       ts,
				ctx:      ctx,
				tableDef: tableDef,
				blks:     blks[(i-1)*step:],
				sels:     make([]int64, 0, 1024),
			}
		} else {
			readers[i] = &blockMergeReader{
				fs:       fs,
				ts:       ts,
				ctx:      ctx,
				tableDef: tableDef,
				blks:     blks[(i-1)*step : i*step],
				sels:     make([]int64, 0, 1024),
			}
		}
	}
	return readers, nil
}

func comparePartitionIndexEntry(a, b PartitionIndexEntry) bool {
	return a.Tuple.Less(b.Tuple)
}

func (p *Partition) Exists(ts timestamp.Timestamp, rowid types.Rowid) bool {
	//TODO
	return false
}

func (p *Partition) addDeleteBatch(batch *batch.Batch) {
	//TODO
	//txID := uuid.NewString()

	//iter := memorytable.NewBatchIter(bat)
	//for {
	//	tuple := iter()
	//	if len(tuple) == 0 {
	//		break
	//	}

	//	rowID := RowID(tuple[0].Value.(types.Rowid))
	//	ts := tuple[1].Value.(types.TS)
	//	t := memtable.Time{
	//		Timestamp: timestamp.Timestamp{
	//			PhysicalTime: ts.Physical(),
	//			LogicalTime:  ts.Logical(),
	//		},
	//	}
	//	tx := memtable.NewTransaction(txID, t, memtable.SnapshotIsolation)

	//	// indexes
	//	var indexes []memtable.Tuple
	//	// block id, time, op
	//	indexes = append(indexes, memtable.Tuple{
	//		index_BlockID_Time_OP,
	//		memtable.ToOrdered(rowIDToBlockID(rowID)),
	//		ts,
	//		memtable.Uint(opDelete),
	//	})
	//	// time
	//	indexes = append(indexes, memtable.Tuple{
	//		index_Time,
	//		ts,
	//	})

	//	err := p.data.Upsert(tx, &DataRow{
	//		rowID: rowID,
	//		value: DataValue{
	//			op: opDelete,
	//		},
	//		indexes: indexes,
	//	})
	//	// the reason to ignore, see comments in Insert method
	//	if moerr.IsMoErrCode(err, moerr.ErrTxnWriteConflict) {
	//		continue
	//	}
	//	if err != nil {
	//		return err
	//	}

	//	if err := tx.Commit(t); err != nil {
	//		return err
	//	}
	//}

	//return nil
}

func (p *Partition) addInsertBatch(batch *batch.Batch) {
	//TODO
	//txID := uuid.NewString()

	//iter := memorytable.NewBatchIter(bat)
	//for {
	//	tuple := iter()
	//	if len(tuple) == 0 {
	//		break
	//	}

	//	rowID := RowID(tuple[0].Value.(types.Rowid))
	//	ts := tuple[1].Value.(types.TS)
	//	t := memtable.Time{
	//		Timestamp: timestamp.Timestamp{
	//			PhysicalTime: ts.Physical(),
	//			LogicalTime:  ts.Logical(),
	//		},
	//	}
	//	tx := memtable.NewTransaction(txID, t, memtable.SnapshotIsolation)

	//	// check primary key
	//	var primaryKey any
	//	if primaryKeyIndex >= 0 {
	//		primaryKey = memtable.ToOrdered(tuple[primaryKeyIndex].Value)
	//		if needCheck && p.cache.PrimaryKeyExists(primaryKey, ts.ToTimestamp()) {
	//			return moerr.NewDuplicate(ctx)
	//		}
	//	}

	//	dataValue := DataValue{
	//		op:    opInsert,
	//		value: make(map[string]memtable.Nullable),
	//	}
	//	for i := 2; i < len(tuple); i++ {
	//		dataValue.value[bat.Attrs[i]] = tuple[i]
	//	}

	//	// indexes
	//	var indexes []memtable.Tuple
	//	// primary key
	//	if primaryKey != nil {
	//		indexes = append(indexes, memtable.Tuple{
	//			index_PrimaryKey,
	//			primaryKey,
	//		})
	//	}
	//	// block id, time, op
	//	indexes = append(indexes, memtable.Tuple{
	//		index_BlockID_Time_OP,
	//		memtable.ToOrdered(rowIDToBlockID(rowID)),
	//		ts,
	//		memtable.Uint(opInsert),
	//	})
	//	// time
	//	indexes = append(indexes, memtable.Tuple{
	//		index_Time,
	//		ts,
	//	})
	//	// columns indexes
	//	for _, def := range p.columnsIndexDefs {
	//		index := memtable.Tuple{
	//			def.Name,
	//		}
	//		for _, col := range def.Columns {
	//			index = append(index, memtable.ToOrdered(tuple[col].Value))
	//		}
	//		indexes = append(indexes, index)
	//	}

	//	err = p.data.Upsert(tx, &DataRow{
	//		rowID:   rowID,
	//		value:   dataValue,
	//		indexes: indexes,
	//	})
	//	// if conflict comes up here,  probably the checkpoint from dn
	//	// has duplicated history versions. As txn write conflict has been
	//	// checked in dn, so it is safe to ignore this error
	//	if moerr.IsMoErrCode(err, moerr.ErrTxnWriteConflict) {
	//		continue
	//	}
	//	if err != nil {
	//		return err
	//	}
	//	if err := tx.Commit(t); err != nil {
	//		return err
	//	}
	//}

	//return nil
}

func (p *Partition) PrimaryKeyExists(ts timestamp.Timestamp, key any) bool {
	//TODO
	return false
}

func (p *Partition) Truncate(ts timestamp.Timestamp) {
	//TODO
}

func (p *Partition) GetRowsByIndex(
	ts timestamp.Timestamp,
	index memorytable.Tuple,
	columns []string,
	deletes map[types.Rowid]uint8,
) (
	rows [][]any,
	err error,
) {
	//TODO
	return
	//t := memtable.Time{
	//	Timestamp: ts,
	//}
	//tx := memtable.NewTransaction(
	//	uuid.NewString(),
	//	t,
	//	memtable.SnapshotIsolation,
	//)
	//iter := p.data.NewIndexIter(tx, index, index)
	//for ok := iter.First(); ok; ok = iter.Next() {
	//	entry := iter.Item()
	//	if _, ok := deletes[types.Rowid(entry.Key)]; ok {
	//		continue
	//	}
	//	data, err := p.data.Get(tx, entry.Key)
	//	if err != nil {
	//		return nil, err
	//	}
	//	rows = append(rows, genRow(&data, columns))
	//}
	//return
}

func (p *Partition) GetRowsByIndexPrefix(
	ts timestamp.Timestamp,
	prefix memorytable.Tuple,
) (
	rows []DataValue,
	err error,
) {
	//TODO
	return
	//t := memtable.Time{
	//	Timestamp: ts,
	//}
	//tx := memtable.NewTransaction(
	//	uuid.NewString(),
	//	t,
	//	memtable.SnapshotIsolation,
	//)
	//iter := p.data.NewIndexIter(
	//	tx,
	//	append(append(prefix[:0:0], prefix...), memtable.Min),
	//	append(append(prefix[:0:0], prefix...), memtable.Max),
	//)
	//for ok := iter.First(); ok; ok = iter.Next() {
	//	entry := iter.Item()
	//	data, err := p.data.Get(tx, entry.Key)
	//	if err != nil {
	//		return nil, err
	//	}
	//	rows = append(rows, data)
	//}
}

func (p *Partition) DeleteByBlockID(
	ctx context.Context,
	ts timestamp.Timestamp,
	blockID uint64,
) (
	err error,
) {
	//TODO
	return nil
	//tx := memtable.NewTransaction(uuid.NewString(), memtable.Time{
	//	Timestamp: ts,
	//}, memtable.SnapshotIsolation)
	//min := memtable.Tuple{
	//	index_BlockID_Time_OP,
	//	memtable.ToOrdered(blockID),
	//	memtable.Min,
	//	memtable.Uint(opInsert),
	//}
	//max := memtable.Tuple{
	//	index_BlockID_Time_OP,
	//	memtable.ToOrdered(blockID),
	//	memtable.Max,
	//	memtable.Uint(opInsert),
	//}
	//iter := p.data.NewIndexIter(tx, min, max)
	//defer iter.Close()
	//for ok := iter.First(); ok; ok = iter.Next() {
	//	entry := iter.Item()
	//	if err := p.data.Delete(tx, entry.Key); err != nil {
	//		return err
	//	}
	//}
	//return tx.Commit(tx.Time)
}

func (p *Partition) IterDeletedRowIDs(
	ctx context.Context,
	blockIDs []uint64,
	ts timestamp.Timestamp,
	fn func(rowID types.Rowid) bool,
) {
	//TODO
	//tx := memtable.NewTransaction(uuid.NewString(), memtable.Time{
	//	Timestamp: ts,
	//}, memtable.SnapshotIsolation)

	//for _, blockID := range blockIDs {
	//	min := memtable.Tuple{
	//		index_BlockID_Time_OP,
	//		memtable.ToOrdered(blockID),
	//		memtable.Min,
	//		memtable.Min,
	//	}
	//	max := memtable.Tuple{
	//		index_BlockID_Time_OP,
	//		memtable.ToOrdered(blockID),
	//		types.TimestampToTS(ts),
	//		memtable.Max,
	//	}
	//	iter := p.data.NewIndexIter(tx, min, max)
	//	defer iter.Close()
	//	deleted := make(map[RowID]bool)
	//	inserted := make(map[RowID]bool)
	//	for ok := iter.First(); ok; ok = iter.Next() {
	//		entry := iter.Item()
	//		rowID := entry.Key
	//		switch entry.Index[3].(memtable.Uint) {
	//		case opInsert:
	//			inserted[rowID] = true
	//		case opDelete:
	//			deleted[rowID] = true
	//		}
	//	}
	//	for rowID := range deleted {
	//		if !inserted[rowID] {
	//			if !fn(rowID) {
	//				break
	//			}
	//		}
	//	}
	//}
}

func (p *Partition) NumRows(
	tx *memtable.Transaction,
	deletes map[types.Rowid]uint8,
	skipBlocks map[uint64]uint8,
) (
	rows int64,
	err error,
) {
	//TODO
	return
	//var rows int64 = 0
	//iter := p.data.NewIter(tx)
	//defer iter.Close()
	//for ok := iter.First(); ok; ok = iter.Next() {
	//	dataKey, dataValue, err := iter.Read()
	//	if err != nil {
	//		return 0, err
	//	}

	//	if _, ok := deletes[types.Rowid(dataKey)]; ok {
	//		continue
	//	}

	//	if dataValue.op == opDelete {
	//		continue
	//	}

	//	if skipBlocks != nil {
	//		if _, ok := skipBlocks[rowIDToBlockID(dataKey)]; ok {
	//			continue
	//		}
	//	}
	//	rows++
	//}
}

func (p *Partition) NewIter(
	ts timestamp.Timestamp,
	min memorytable.Tuple,
	max memorytable.Tuple,
) *PartitionIter {
	return &PartitionIter{}
}

func (p *PartitionIter) First() bool {
	//TODO
	return false
}

func (p *PartitionIter) Next() bool {
	//TODO
	return false
}

func (p *PartitionIter) Item() (rowid types.Rowid, row DataValue) {
	//TODO
	return
}

func (p *PartitionIter) Close() error {
	//TODO
	return nil
}
