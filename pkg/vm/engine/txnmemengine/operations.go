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
	"encoding/gob"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

const (
	opCreateDatabase = iota + 64
	opOpenDatabase
	opGetDatabases
	opDeleteDatabase
	opCreateRelation
	opDeleteRelation
	opOpenRelation
	opGetRelations
	opAddTableDef
	opDelTableDef
	opDelete
	opGetPrimaryKeys
	opGetTableDefs
	opTruncate
	opUpdate
	opWrite
	opNewTableIter
	opRead
	opCloseTableIter
)

func mustEncodePayload(o any) []byte {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(o); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

type createDatabaseReq struct {
	Name string
}

type openDatabaseReq struct {
	Name string
}

type openDatabaseResp struct {
	ID int64
}

type getDatabasesResp struct {
	Names []string
}

type deleteDatabaseReq struct {
	Name string
}

type createRelationReq struct {
	DatabaseID int64
	Name       string
	Defs       []engine.TableDef
}

type deleteRelationReq struct {
	DatabaseID int64
	Name       string
}

type openRelationReq struct {
	DatabaseID int64
	Name       string
}

type openRelationResp struct {
	ID   int64
	Type RelationType
}

type getRelationsReq struct {
	DatabaseID int64
}

type getRelationsResp struct {
	Names []string
}

type addTableDefReq struct {
	TableID int64
	Def     engine.TableDef
}

type delTableDefReq struct {
	TableID int64
	Def     engine.TableDef
}

type deleteReq struct {
	TableID int64
	Vector  *vector.Vector
}

type getPrimaryKeysReq struct {
	TableID int64
}

type getPrimaryKeysResp struct {
	Attrs []*engine.Attribute
}

type getTableDefsReq struct {
	TableID int64
}

type getTableDefsResp struct {
	Defs []engine.TableDef
}

type truncateReq struct {
	TableID int64
}

type truncateResp struct {
	AffectedRows int64
}

type updateReq struct {
	TableID int64
	Batch   *batch.Batch
}

type writeReq struct {
	TableID int64
	Batch   *batch.Batch
}

type newTableIterReq struct {
	TableID int64
	Expr    *plan.Expr
	Data    []byte
}

type newTableIterResp struct {
	IterID int64
}

type readReq struct {
	IterID   int64
	ColNames []string
}

type readResp struct {
	Batch *batch.Batch
}

type closeTableIterReq struct {
	IterID int64
}
