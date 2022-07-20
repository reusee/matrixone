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
		return handleRead(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.OpenDatabaseReq,
				resp *txnmemengine.OpenDatabaseResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpGetDatabases:
		return handleReadNoReq(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				resp *txnmemengine.OpenDatabaseResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpOpenRelation:
		return handleRead(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.OpenRelationReq,
				resp *txnmemengine.OpenRelationResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpGetRelations:
		return handleRead(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.GetRelationsReq,
				resp *txnmemengine.GetRelationsResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpGetPrimaryKeys:
		return handleRead(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.GetPrimaryKeysReq,
				resp *txnmemengine.GetPrimaryKeysResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpGetTableDefs:
		return handleRead(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.GetTableDefsReq,
				resp *txnmemengine.GetTableDefsResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpNewTableIter:
		return handleRead(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.NewTableIterReq,
				resp *txnmemengine.NewTableIterResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpRead:
		return handleRead(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.ReadReq,
				resp *txnmemengine.ReadResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpCloseTableIter:
		return handleReadNoResp(
			s, txnMeta, payload,
			func(
				tx *Transaction,
				req txnmemengine.CloseTableIterReq,
			) error {
				//TODO
				return nil
			},
		)

	}

	panic("bad op")
}

func handleRead[
	Req any,
	Resp any,
](
	s *Storage,
	txnMeta txn.TxnMeta,
	payload []byte,
	fn func(
		*Transaction,
		Req,
		*Resp,
	) error,
) (
	res storage.ReadResult,
	err error,
) {

	var req Req
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&req); err != nil {
		return nil, err
	}

	var resp Resp
	tx := s.getTransaction(txnMeta)
	if err := fn(tx, req, &resp); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(resp); err != nil {
		return nil, err
	}

	return &readResult{
		payload: buf.Bytes(),
	}, nil
}

func handleReadNoReq[
	Resp any,
](
	s *Storage,
	txnMeta txn.TxnMeta,
	payload []byte,
	fn func(
		*Transaction,
		*Resp,
	) error,
) (
	res storage.ReadResult,
	err error,
) {

	var resp Resp
	tx := s.getTransaction(txnMeta)
	if err := fn(tx, &resp); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(resp); err != nil {
		return nil, err
	}

	return &readResult{
		payload: buf.Bytes(),
	}, nil
}

func handleReadNoResp[
	Req any,
](
	s *Storage,
	txnMeta txn.TxnMeta,
	payload []byte,
	fn func(
		*Transaction,
		Req,
	) error,
) (
	res storage.ReadResult,
	err error,
) {

	var req Req
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&req); err != nil {
		return nil, err
	}

	tx := s.getTransaction(txnMeta)
	if err := fn(tx, req); err != nil {
		return nil, err
	}

	return &readResult{
		payload: nil,
	}, nil
}
