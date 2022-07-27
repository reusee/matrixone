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
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnengine"
	"github.com/stretchr/testify/assert"
)

func testDatabase(
	t *testing.T,
	newStorage func() (*Storage, error),
) {
	s, err := newStorage()
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

	{
		resp := testRead[txnengine.OpenDatabaseResp](
			t, s, txnMeta,
			txnengine.OpOpenDatabase,
			txnengine.OpenDatabaseReq{
				Name: "foo",
			},
		)
		assert.Equal(t, true, resp.ErrNotFound)
	}

	{
		resp := testWrite[txnengine.CreateDatabaseResp](
			t, s, txnMeta,
			txnengine.OpCreateDatabase,
			txnengine.CreateDatabaseReq{
				Name: "foo",
			},
		)
		assert.Equal(t, false, resp.ErrExisted)
	}

	{
		resp := testRead[txnengine.OpenDatabaseResp](
			t, s, txnMeta,
			txnengine.OpOpenDatabase,
			txnengine.OpenDatabaseReq{
				Name: "foo",
			},
		)
		assert.Equal(t, false, resp.ErrNotFound)
		assert.NotNil(t, resp.ID)
	}

	{
		resp := testRead[txnengine.GetDatabasesResp](
			t, s, txnMeta,
			txnengine.OpGetDatabases,
			txnengine.GetDatabasesReq{},
		)
		assert.Equal(t, 1, len(resp.Names))
		assert.Equal(t, "foo", resp.Names[0])
	}

}

func testRead[
	Resp any,
	Req any,
](
	t *testing.T,
	s *Storage,
	txnMeta txn.TxnMeta,
	op uint32,
	req Req,
) (
	resp Resp,
) {

	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(req)
	assert.Nil(t, err)

	res, err := s.Read(txnMeta, op, buf.Bytes())
	assert.Nil(t, err)
	data, err := res.Read()
	assert.Nil(t, err)

	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&resp)
	assert.Nil(t, err)

	return
}

func testWrite[
	Resp any,
	Req any,
](
	t *testing.T,
	s *Storage,
	txnMeta txn.TxnMeta,
	op uint32,
	req Req,
) (
	resp Resp,
) {

	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(req)
	assert.Nil(t, err)

	data, err := s.Write(txnMeta, op, buf.Bytes())
	assert.Nil(t, err)

	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&resp)
	assert.Nil(t, err)

	return
}
