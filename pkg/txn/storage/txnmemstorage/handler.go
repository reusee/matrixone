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

type Handler interface {
	HandleOpenDatabase(
		meta txn.TxnMeta,
		req txnmemengine.OpenDatabaseReq,
		resp *txnmemengine.OpenDatabaseResp,
	) error

	HandleGetDatabases(
		meta txn.TxnMeta,
		req txnmemengine.GetDatabasesReq,
		resp *txnmemengine.GetDatabasesResp,
	) error

	HandleOpenRelation(
		meta txn.TxnMeta,
		req txnmemengine.OpenRelationReq,
		resp *txnmemengine.OpenRelationResp,
	) error

	HandleGetRelations(
		meta txn.TxnMeta,
		req txnmemengine.GetRelationsReq,
		resp *txnmemengine.GetRelationsResp,
	) error

	HandleGetPrimaryKeys(
		meta txn.TxnMeta,
		req txnmemengine.GetPrimaryKeysReq,
		resp *txnmemengine.GetPrimaryKeysResp,
	) error

	HandleGetTableDefs(
		meta txn.TxnMeta,
		req txnmemengine.GetTableDefsReq,
		resp *txnmemengine.GetTableDefsResp,
	) error

	HandleNewTableIter(
		meta txn.TxnMeta,
		req txnmemengine.NewTableIterReq,
		resp *txnmemengine.NewTableIterResp,
	) error

	HandleRead(
		meta txn.TxnMeta,
		req txnmemengine.ReadReq,
		resp *txnmemengine.ReadResp,
	) error

	HandleCloseTableIter(
		meta txn.TxnMeta,
		req txnmemengine.CloseTableIterReq,
		resp *txnmemengine.CloseTableIterResp,
	) error

	HandleCreateDatabase(
		meta txn.TxnMeta,
		req txnmemengine.CreateDatabaseReq,
		resp *txnmemengine.CreateDatabaseResp,
	) error

	HandleDeleteDatabase(
		meta txn.TxnMeta,
		req txnmemengine.DeleteDatabaseReq,
		resp *txnmemengine.DeleteDatabaseResp,
	) error

	HandleCreateRelation(
		meta txn.TxnMeta,
		req txnmemengine.CreateRelationReq,
		resp *txnmemengine.CreateRelationResp,
	) error

	HandleDeleteRelation(
		meta txn.TxnMeta,
		req txnmemengine.DeleteRelationReq,
		resp *txnmemengine.DeleteRelationResp,
	) error

	HandleAddTableDef(
		meta txn.TxnMeta,
		req txnmemengine.AddTableDefReq,
		resp *txnmemengine.AddTableDefResp,
	) error

	HandleDelTableDef(
		meta txn.TxnMeta,
		req txnmemengine.DelTableDefReq,
		resp *txnmemengine.DelTableDefResp,
	) error

	HandleDelete(
		meta txn.TxnMeta,
		req txnmemengine.DeleteReq,
		resp *txnmemengine.DeleteResp,
	) error

	HandleTruncate(
		meta txn.TxnMeta,
		req txnmemengine.TruncateReq,
		resp *txnmemengine.TruncateResp,
	) error

	HandleUpdate(
		meta txn.TxnMeta,
		req txnmemengine.UpdateReq,
		resp *txnmemengine.UpdateResp,
	) error

	HandleWrite(
		meta txn.TxnMeta,
		req txnmemengine.WriteReq,
		resp *txnmemengine.WriteResp,
	) error
}
