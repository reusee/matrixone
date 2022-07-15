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
