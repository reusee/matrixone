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

package txnmemengine

import (
	"bytes"
	"encoding/gob"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

const (
	OpCreateDatabase = iota + 64
	OpOpenDatabase
	OpGetDatabases
	OpDeleteDatabase
	OpCreateRelation
	OpDeleteRelation
	OpOpenRelation
	OpGetRelations
	OpAddTableDef
	OpDelTableDef
	OpDelete
	OpGetPrimaryKeys
	OpGetTableDefs
	OpTruncate
	OpUpdate
	OpWrite
	OpNewTableIter
	OpRead
	OpCloseTableIter
)

func mustEncodePayload(o any) []byte {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(o); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

type CreateDatabaseReq struct {
	Name string
}

type OpenDatabaseReq struct {
	Name string
}

type OpenDatabaseResp struct {
	ID int64
}

type GetDatabasesResp struct {
	Names []string
}

type DeleteDatabaseReq struct {
	Name string
}

type CreateRelationReq struct {
	DatabaseID int64
	Name       string
	Defs       []engine.TableDef
}

type DeleteRelationReq struct {
	DatabaseID int64
	Name       string
}

type OpenRelationReq struct {
	DatabaseID int64
	Name       string
}

type OpenRelationResp struct {
	ID   int64
	Type RelationType
}

type GetRelationsReq struct {
	DatabaseID int64
}

type GetRelationsResp struct {
	Names []string
}

type AddTableDefReq struct {
	TableID int64
	Def     engine.TableDef
}

type DelTableDefReq struct {
	TableID int64
	Def     engine.TableDef
}

type DeleteReq struct {
	TableID int64
	Vector  *vector.Vector
}

type GetPrimaryKeysReq struct {
	TableID int64
}

type GetPrimaryKeysResp struct {
	Attrs []*engine.Attribute
}

type GetTableDefsReq struct {
	TableID int64
}

type GetTableDefsResp struct {
	Defs []engine.TableDef
}

type TruncateReq struct {
	TableID int64
}

type TruncateResp struct {
	AffectedRows int64
}

type UpdateReq struct {
	TableID int64
	Batch   *batch.Batch
}

type WriteReq struct {
	TableID int64
	Batch   *batch.Batch
}

type NewTableIterReq struct {
	TableID int64
	Expr    *plan.Expr
	Shards  [][]byte
}

type NewTableIterResp struct {
	IterID int64
}

type ReadReq struct {
	IterID   int64
	ColNames []string
}

type ReadResp struct {
	Batch *batch.Batch
}

type CloseTableIterReq struct {
	IterID int64
}
