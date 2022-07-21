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

	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnmemengine"
)

func (s *Storage) Write(txnMeta txn.TxnMeta, op uint32, payload []byte) (result []byte, err error) {

	switch op {

	case txnmemengine.OpCreateDatabase:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleCreateDatabase,
		)

	case txnmemengine.OpDeleteDatabase:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleDeleteDatabase,
		)

	case txnmemengine.OpCreateRelation:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleCreateRelation,
		)

	case txnmemengine.OpDeleteRelation:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleDeleteRelation,
		)

	case txnmemengine.OpAddTableDef:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleAddTableDef,
		)

	case txnmemengine.OpDelTableDef:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleDelTableDef,
		)

	case txnmemengine.OpDelete:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleDelete,
		)

	case txnmemengine.OpTruncate:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleTruncate,
		)

	case txnmemengine.OpUpdate:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleUpdate,
		)

	case txnmemengine.OpWrite:
		return handleWrite(
			s, txnMeta, payload,
			s.handler.HandleWrite,
		)

	}

	return
}

func handleWrite[
	Req any,
	Resp any,
](
	s *Storage,
	meta txn.TxnMeta,
	payload []byte,
	fn func(
		meta txn.TxnMeta,
		req Req,
		resp *Resp,
	) (
		err error,
	),
) (
	res []byte,
	err error,
) {

	var req Req
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&req); err != nil {
		return nil, err
	}

	var resp Resp
	err = fn(meta, req, &resp)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(resp); err != nil {
		return nil, err
	}
	res = buf.Bytes()

	return
}
