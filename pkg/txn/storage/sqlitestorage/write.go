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

	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnmemengine"
)

func (s *Storage) Write(txnMeta txn.TxnMeta, op uint32, payload []byte) (result []byte, err error) {

	switch op {

	case txnmemengine.OpCreateDatabase:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.CreateDatabaseReq,
				resp *txnmemengine.CreateDatabaseResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpDeleteDatabase:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.DeleteDatabaseReq,
				resp *txnmemengine.DeleteDatabaseResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpCreateRelation:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.CreateRelationReq,
				resp *txnmemengine.CreateRelationResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpDeleteRelation:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.DeleteRelationReq,
				resp *txnmemengine.DeleteRelationResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpAddTableDef:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.AddTableDefReq,
				resp *txnmemengine.AddTableDefResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpDelTableDef:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.DelTableDefReq,
				resp *txnmemengine.DelTableDefResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpDelete:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.DeleteReq,
				resp *txnmemengine.DeleteResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpTruncate:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.TruncateReq,
				resp *txnmemengine.TruncateResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpUpdate:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.UpdateReq,
				resp *txnmemengine.UpdateResp,
			) error {
				//TODO
				return nil
			},
		)

	case txnmemengine.OpWrite:
		return handleWrite(
			s, txnMeta, payload,
			func(
				req txnmemengine.WriteReq,
				resp *txnmemengine.WriteResp,
			) error {
				//TODO
				return nil
			},
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
	err = fn(req, &resp)
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
