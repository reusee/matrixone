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

package engine

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

type Table struct {
	engine      *Engine
	txnOperator client.TxnOperator
	id          int64
}

var _ engine.Relation = new(Table)

func (*Table) Rows() int64 {
	return 1
}

func (*Table) Size(string) int64 {
	return 0
}

func (t *Table) AddTableDef(ctx context.Context, def engine.TableDef) error {

	_, err := doTxnRequest(
		ctx,
		t.txnOperator.Write,
		t.engine.getDataNodes(),
		txn.TxnMethod_Write,
		opAddTableDef,
		addTableDefReq{
			TableID: t.id,
			Def:     def,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (t *Table) DelTableDef(ctx context.Context, def engine.TableDef) error {

	_, err := doTxnRequest(
		ctx,
		t.txnOperator.Write,
		t.engine.getDataNodes(),
		txn.TxnMethod_Write,
		opDelTableDef,
		delTableDefReq{
			TableID: t.id,
			Def:     def,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (t *Table) Delete(ctx context.Context, vec *vector.Vector, _ string) error {

	shards, err := t.engine.shardPolicy.Vector(vec, t.engine.getDataNodes())
	if err != nil {
		return err
	}

	for _, shard := range shards {
		_, err := doTxnRequest(
			ctx,
			t.txnOperator.Write,
			shard.Nodes,
			txn.TxnMethod_Write,
			opDelete,
			deleteReq{
				TableID: t.id,
				Vector:  shard.Vector,
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (*Table) GetHideKey() *engine.Attribute {
	return nil
}

func (*Table) GetPriKeyOrHideKey() ([]engine.Attribute, bool) {
	return nil, false
}

func (t *Table) GetPrimaryKeys(ctx context.Context) ([]*engine.Attribute, error) {

	resps, err := doTxnRequest(
		ctx,
		t.txnOperator.Read,
		t.engine.getDataNodes()[:1],
		txn.TxnMethod_Read,
		opGetPrimaryKeys,
		getPrimaryKeysReq{
			TableID: t.id,
		},
	)
	if err != nil {
		return nil, err
	}

	var resp getPrimaryKeysResp
	if err := gob.NewDecoder(bytes.NewReader(resps[0])).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Attrs, nil
}

func (*Table) NewReader(ctx context.Context, parallel int, expr *plan.Expr, data []byte) []engine.Reader {
	//TODO
	return nil
}

func (t *Table) Nodes() engine.Nodes {
	return t.engine.Nodes()
}

func (t *Table) TableDefs(ctx context.Context) ([]engine.TableDef, error) {

	resps, err := doTxnRequest(
		ctx,
		t.txnOperator.Read,
		t.engine.getDataNodes()[:1],
		txn.TxnMethod_Read,
		opGetTableDefs,
		getTableDefsReq{
			TableID: t.id,
		},
	)
	if err != nil {
		return nil, err
	}

	var resp getTableDefsResp
	if err := gob.NewDecoder(bytes.NewReader(resps[0])).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Defs, nil
}

func (t *Table) Truncate(ctx context.Context) (uint64, error) {

	resps, err := doTxnRequest(
		ctx,
		t.txnOperator.Write,
		t.engine.getDataNodes(),
		txn.TxnMethod_Write,
		opTruncate,
		truncateReq{
			TableID: t.id,
		},
	)
	if err != nil {
		return 0, err
	}

	var affectedRows int64
	for _, payload := range resps {
		var r truncateResp
		if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&r); err != nil {
			return uint64(affectedRows), err
		}
		affectedRows += r.AffectedRows
	}

	return uint64(affectedRows), nil
}

func (t *Table) Update(ctx context.Context, data *batch.Batch) error {

	shards, err := t.engine.shardPolicy.Batch(data, t.engine.getDataNodes())
	if err != nil {
		return err
	}

	for _, shard := range shards {
		_, err := doTxnRequest(
			ctx,
			t.txnOperator.Write,
			shard.Nodes,
			txn.TxnMethod_Write,
			opUpdate,
			updateReq{
				TableID: t.id,
				Batch:   shard.Batch,
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Table) Write(ctx context.Context, data *batch.Batch) error {

	shards, err := t.engine.shardPolicy.Batch(data, t.engine.getDataNodes())
	if err != nil {
		return err
	}

	for _, shard := range shards {
		_, err := doTxnRequest(
			ctx,
			t.txnOperator.Write,
			shard.Nodes,
			txn.TxnMethod_Write,
			opWrite,
			writeReq{
				TableID: t.id,
				Batch:   shard.Batch,
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
