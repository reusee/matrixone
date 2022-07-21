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

package sqlitestorage

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnmemengine"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	s, err := New()
	assert.Nil(t, err)
	defer s.Close()

	txnMeta := txn.TxnMeta{
		ID:     []byte("1"),
		Status: txn.TxnStatus_Active,
		SnapshotTS: timestamp.Timestamp{
			PhysicalTime: 1,
			LogicalTime:  1,
		},
	}

	req := txnmemengine.OpenDatabaseReq{
		Name: "foo",
	}
	buf := new(bytes.Buffer)
	err = gob.NewEncoder(buf).Encode(req)
	assert.Nil(t, err)

	res, err := s.Read(txnMeta, txnmemengine.OpOpenDatabase, buf.Bytes())
	assert.Nil(t, err)
	data, err := res.Read()
	assert.Nil(t, err)

	var resp txnmemengine.OpenDatabaseResp
	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&resp)
	assert.Nil(t, err)

	assert.Equal(t, true, resp.ErrNotFound)

}
