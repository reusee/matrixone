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
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnmemengine"
)

func (s *Storage) Write(txnMeta txn.TxnMeta, op uint32, payload []byte) ([]byte, error) {

	switch op {

	case txnmemengine.OpCreateDatabase:
		return handleWriteNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.CreateDatabaseReq,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpDeleteDatabase:
		return handleWriteNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.DeleteDatabaseReq,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpCreateRelation:
		return handleWriteNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.CreateRelationReq,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpDeleteRelation:
		return handleWriteNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.DeleteRelationReq,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpAddTableDef:
		return handleWriteNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.AddTableDefReq,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpDelTableDef:
		return handleWriteNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.DelTableDefReq,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpDelete:
		return handleWriteNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.DeleteReq,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpTruncate:
		return handleWrite(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.TruncateReq,
				resp *txnmemengine.TruncateResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpUpdate:
		return handleWriteNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.UpdateReq,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpWrite:
		return handleWriteNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.WriteReq,
			) error {
				//TODO
				return nil
			},
		)

	}

	panic("bad op")
}

func handleWrite[
	Req any,
	Resp any,
](
	s *Storage,
	meta txn.TxnMeta,
	payload []byte,
	fn func(
		tx *Transaction,
		req Req,
		resp *Resp,
	) error,
) (
	res []byte,
	err error,
) {

	var req Req
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&req); err != nil {
		return nil, err
	}

	tx, err := s.getTransaction(meta)
	if err != nil {
		return nil, err
	}

	var resp Resp
	if err := fn(tx, req, &resp); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(resp); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func handleWriteNoResp[
	Req any,
](
	s *Storage,
	meta txn.TxnMeta,
	payload []byte,
	fn func(
		tx *Transaction,
		req Req,
	) error,
) (
	res []byte,
	err error,
) {

	var req Req
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&req); err != nil {
		return nil, err
	}

	tx, err := s.getTransaction(meta)
	if err != nil {
		return nil, err
	}

	if err := fn(tx, req); err != nil {
		return nil, err
	}

	return nil, nil
}
