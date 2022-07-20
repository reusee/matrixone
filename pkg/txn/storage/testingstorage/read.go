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

package storage

import (
	"bytes"
	"encoding/gob"

	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/storage"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnmemengine"
)

func (s *Storage) Read(txnMeta txn.TxnMeta, op uint32, payload []byte) (storage.ReadResult, error) {
	switch op {

	case txnmemengine.OpOpenDatabase:
		var req txnmemengine.OpenDatabaseReq
		if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&req); err != nil {
			return nil, err
		}
		var resp txnmemengine.OpenDatabaseResp
		tx := s.getTransaction(txnMeta)
		_ = tx
		resp.ID = 42 //TODO
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(resp); err != nil {
			return nil, err
		}
		return &readResult{
			payload: buf.Bytes(),
		}, nil

	case txnmemengine.OpGetDatabases:
		//TODO

	case txnmemengine.OpOpenRelation:
		//TODO

	case txnmemengine.OpGetRelations:
		//TODO

	case txnmemengine.OpGetPrimaryKeys:
		//TODO

	case txnmemengine.OpGetTableDefs:
		//TODO

	case txnmemengine.OpNewTableIter:
		//TODO

	case txnmemengine.OpRead:
		//TODO

	case txnmemengine.OpCloseTableIter:
		//TODO

	}

	panic("bad op")
}
