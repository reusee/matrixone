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
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnmemengine"
)

type MemHandler struct {
}

func NewMemHandler() *MemHandler {
	h := &MemHandler{}
	return h
}

var _ Handler = new(MemHandler)

// HandleAddTableDef implements Handler
func (*MemHandler) HandleAddTableDef(meta txn.TxnMeta, req txnmemengine.AddTableDefReq, resp *txnmemengine.AddTableDefResp) error {
	panic("unimplemented")
}

// HandleCloseTableIter implements Handler
func (*MemHandler) HandleCloseTableIter(meta txn.TxnMeta, req txnmemengine.CloseTableIterReq, resp *txnmemengine.CloseTableIterResp) error {
	panic("unimplemented")
}

// HandleCreateDatabase implements Handler
func (*MemHandler) HandleCreateDatabase(meta txn.TxnMeta, req txnmemengine.CreateDatabaseReq, resp *txnmemengine.CreateDatabaseResp) error {
	panic("unimplemented")
}

// HandleCreateRelation implements Handler
func (*MemHandler) HandleCreateRelation(meta txn.TxnMeta, req txnmemengine.CreateRelationReq, resp *txnmemengine.CreateRelationResp) error {
	panic("unimplemented")
}

// HandleDelTableDef implements Handler
func (*MemHandler) HandleDelTableDef(meta txn.TxnMeta, req txnmemengine.DelTableDefReq, resp *txnmemengine.DelTableDefResp) error {
	panic("unimplemented")
}

// HandleDelete implements Handler
func (*MemHandler) HandleDelete(meta txn.TxnMeta, req txnmemengine.DeleteReq, resp *txnmemengine.DeleteResp) error {
	panic("unimplemented")
}

// HandleDeleteDatabase implements Handler
func (*MemHandler) HandleDeleteDatabase(meta txn.TxnMeta, req txnmemengine.DeleteDatabaseReq, resp *txnmemengine.DeleteDatabaseResp) error {
	panic("unimplemented")
}

// HandleDeleteRelation implements Handler
func (*MemHandler) HandleDeleteRelation(meta txn.TxnMeta, req txnmemengine.DeleteRelationReq, resp *txnmemengine.DeleteRelationResp) error {
	panic("unimplemented")
}

// HandleGetDatabases implements Handler
func (*MemHandler) HandleGetDatabases(meta txn.TxnMeta, req txnmemengine.GetDatabasesReq, resp *txnmemengine.GetDatabasesResp) error {
	panic("unimplemented")
}

// HandleGetPrimaryKeys implements Handler
func (*MemHandler) HandleGetPrimaryKeys(meta txn.TxnMeta, req txnmemengine.GetPrimaryKeysReq, resp *txnmemengine.GetPrimaryKeysResp) error {
	panic("unimplemented")
}

// HandleGetRelations implements Handler
func (*MemHandler) HandleGetRelations(meta txn.TxnMeta, req txnmemengine.GetRelationsReq, resp *txnmemengine.GetRelationsResp) error {
	panic("unimplemented")
}

// HandleGetTableDefs implements Handler
func (*MemHandler) HandleGetTableDefs(meta txn.TxnMeta, req txnmemengine.GetTableDefsReq, resp *txnmemengine.GetTableDefsResp) error {
	panic("unimplemented")
}

// HandleNewTableIter implements Handler
func (*MemHandler) HandleNewTableIter(meta txn.TxnMeta, req txnmemengine.NewTableIterReq, resp *txnmemengine.NewTableIterResp) error {
	panic("unimplemented")
}

// HandleOpenDatabase implements Handler
func (*MemHandler) HandleOpenDatabase(meta txn.TxnMeta, req txnmemengine.OpenDatabaseReq, resp *txnmemengine.OpenDatabaseResp) error {
	panic("unimplemented")
}

// HandleOpenRelation implements Handler
func (*MemHandler) HandleOpenRelation(meta txn.TxnMeta, req txnmemengine.OpenRelationReq, resp *txnmemengine.OpenRelationResp) error {
	panic("unimplemented")
}

// HandleRead implements Handler
func (*MemHandler) HandleRead(meta txn.TxnMeta, req txnmemengine.ReadReq, resp *txnmemengine.ReadResp) error {
	panic("unimplemented")
}

// HandleTruncate implements Handler
func (*MemHandler) HandleTruncate(meta txn.TxnMeta, req txnmemengine.TruncateReq, resp *txnmemengine.TruncateResp) error {
	panic("unimplemented")
}

// HandleUpdate implements Handler
func (*MemHandler) HandleUpdate(meta txn.TxnMeta, req txnmemengine.UpdateReq, resp *txnmemengine.UpdateResp) error {
	panic("unimplemented")
}

// HandleWrite implements Handler
func (*MemHandler) HandleWrite(meta txn.TxnMeta, req txnmemengine.WriteReq, resp *txnmemengine.WriteResp) error {
	panic("unimplemented")
}
