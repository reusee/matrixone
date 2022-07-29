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

	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnengine"
)

type MemHandler struct {
	transactions struct {
		sync.Mutex
		Map map[string]*Transaction
	}
	databases *Table[Text, DatabaseAttrs]
}

func NewMemHandler() *MemHandler {
	h := &MemHandler{}
	h.transactions.Map = make(map[string]*Transaction)
	h.databases = NewTable[Text, DatabaseAttrs]()
	return h
}

var _ Handler = new(MemHandler)

// HandleAddTableDef implements Handler
func (*MemHandler) HandleAddTableDef(meta txn.TxnMeta, req txnengine.AddTableDefReq, resp *txnengine.AddTableDefResp) error {
	//TODO
	panic("unimplemented")
}

// HandleCloseTableIter implements Handler
func (*MemHandler) HandleCloseTableIter(meta txn.TxnMeta, req txnengine.CloseTableIterReq, resp *txnengine.CloseTableIterResp) error {
	//TODO
	panic("unimplemented")
}

// HandleCreateDatabase implements Handler
func (*MemHandler) HandleCreateDatabase(meta txn.TxnMeta, req txnengine.CreateDatabaseReq, resp *txnengine.CreateDatabaseResp) error {
	//TODO
	panic("unimplemented")
}

// HandleCreateRelation implements Handler
func (*MemHandler) HandleCreateRelation(meta txn.TxnMeta, req txnengine.CreateRelationReq, resp *txnengine.CreateRelationResp) error {
	//TODO
	panic("unimplemented")
}

// HandleDelTableDef implements Handler
func (*MemHandler) HandleDelTableDef(meta txn.TxnMeta, req txnengine.DelTableDefReq, resp *txnengine.DelTableDefResp) error {
	//TODO
	panic("unimplemented")
}

// HandleDelete implements Handler
func (*MemHandler) HandleDelete(meta txn.TxnMeta, req txnengine.DeleteReq, resp *txnengine.DeleteResp) error {
	//TODO
	panic("unimplemented")
}

// HandleDeleteDatabase implements Handler
func (*MemHandler) HandleDeleteDatabase(meta txn.TxnMeta, req txnengine.DeleteDatabaseReq, resp *txnengine.DeleteDatabaseResp) error {
	//TODO
	panic("unimplemented")
}

// HandleDeleteRelation implements Handler
func (*MemHandler) HandleDeleteRelation(meta txn.TxnMeta, req txnengine.DeleteRelationReq, resp *txnengine.DeleteRelationResp) error {
	//TODO
	panic("unimplemented")
}

// HandleGetDatabases implements Handler
func (*MemHandler) HandleGetDatabases(meta txn.TxnMeta, req txnengine.GetDatabasesReq, resp *txnengine.GetDatabasesResp) error {
	//TODO
	panic("unimplemented")
}

// HandleGetPrimaryKeys implements Handler
func (*MemHandler) HandleGetPrimaryKeys(meta txn.TxnMeta, req txnengine.GetPrimaryKeysReq, resp *txnengine.GetPrimaryKeysResp) error {
	//TODO
	panic("unimplemented")
}

// HandleGetRelations implements Handler
func (*MemHandler) HandleGetRelations(meta txn.TxnMeta, req txnengine.GetRelationsReq, resp *txnengine.GetRelationsResp) error {
	//TODO
	panic("unimplemented")
}

// HandleGetTableDefs implements Handler
func (*MemHandler) HandleGetTableDefs(meta txn.TxnMeta, req txnengine.GetTableDefsReq, resp *txnengine.GetTableDefsResp) error {
	//TODO
	panic("unimplemented")
}

// HandleNewTableIter implements Handler
func (*MemHandler) HandleNewTableIter(meta txn.TxnMeta, req txnengine.NewTableIterReq, resp *txnengine.NewTableIterResp) error {
	//TODO
	panic("unimplemented")
}

func (m *MemHandler) HandleOpenDatabase(meta txn.TxnMeta, req txnengine.OpenDatabaseReq, resp *txnengine.OpenDatabaseResp) error {
	tx := m.getTx(meta)
	iter := m.databases.NewIter(tx, tx.CurrentTime)
	defer iter.Close()
	for iter.Next() {
		_, attrs := iter.Get()
		if attrs.Name == req.Name {
			resp.ID = attrs.ID
			return nil
		}
	}
	resp.ErrNotFound = true
	return nil
}

// HandleOpenRelation implements Handler
func (*MemHandler) HandleOpenRelation(meta txn.TxnMeta, req txnengine.OpenRelationReq, resp *txnengine.OpenRelationResp) error {
	//TODO
	panic("unimplemented")
}

// HandleRead implements Handler
func (*MemHandler) HandleRead(meta txn.TxnMeta, req txnengine.ReadReq, resp *txnengine.ReadResp) error {
	//TODO
	panic("unimplemented")
}

// HandleTruncate implements Handler
func (*MemHandler) HandleTruncate(meta txn.TxnMeta, req txnengine.TruncateReq, resp *txnengine.TruncateResp) error {
	//TODO
	panic("unimplemented")
}

// HandleUpdate implements Handler
func (*MemHandler) HandleUpdate(meta txn.TxnMeta, req txnengine.UpdateReq, resp *txnengine.UpdateResp) error {
	//TODO
	panic("unimplemented")
}

// HandleWrite implements Handler
func (*MemHandler) HandleWrite(meta txn.TxnMeta, req txnengine.WriteReq, resp *txnengine.WriteResp) error {
	//TODO
	panic("unimplemented")
}

func (m *MemHandler) getTx(meta txn.TxnMeta) *Transaction {
	id := string(meta.ID)
	m.transactions.Lock()
	defer m.transactions.Unlock()
	tx, ok := m.transactions.Map[id]
	if !ok {
		tx = NewTransaction(id, meta.SnapshotTS)
		m.transactions.Map[id] = tx
	}
	return tx
}
