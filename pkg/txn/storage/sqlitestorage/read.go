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
	"database/sql"
	"encoding/gob"
	"errors"

	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/storage"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnmemengine"
)

func (s *Storage) Read(txnMeta txn.TxnMeta, op uint32, payload []byte) (res storage.ReadResult, err error) {

	switch op {

	case txnmemengine.OpOpenDatabase:
		return handleRead(
			s, txnMeta, payload,
			func(
				req txnmemengine.OpenDatabaseReq,
				resp *txnmemengine.OpenDatabaseResp,
			) (
				err error,
			) {

				var args Args
				err = s.db.QueryRow(`
          select id from databases
          where name = `+args.bind(req.Name)+`
          and `+args.visible(txnMeta)+`
          order by min_physical_time desc, min_logical_time desc
          limit 1
          `, args...).Scan(&resp.ID)
				if errors.Is(err, sql.ErrNoRows) {
					err = nil
					resp.ErrNotFound = true
					return
				}
				if err != nil {
					return err
				}

				return nil
			},
		)

	case txnmemengine.OpGetDatabases:
		return handleRead(
			s, txnMeta, payload,
			func(
				req txnmemengine.GetDatabasesReq,
				resp *txnmemengine.GetDatabasesResp,
			) (
				err error,
			) {

				var args Args
				err = queryRows(s.db, `
          select name from databases
          where `+args.visible(txnMeta)+`
        `, func(rows *sql.Rows) error {
					var name string
					if err := rows.Scan(&name); err != nil {
						return err
					}
					resp.Names = append(resp.Names, name)
					return nil
				}, args...)
				if err != nil {
					return err
				}

				return
			},
		)

	case txnmemengine.OpOpenRelation:
		return handleRead(
			s, txnMeta, payload,
			func(
				req txnmemengine.OpenRelationReq,
				resp *txnmemengine.OpenRelationResp,
			) (
				err error,
			) {
				//TODO
				return
			},
		)

	case txnmemengine.OpGetRelations:
		return handleRead(
			s, txnMeta, payload,
			func(
				req txnmemengine.GetRelationsReq,
				resp *txnmemengine.GetRelationsResp,
			) (
				err error,
			) {
				//TODO
				return
			},
		)

	case txnmemengine.OpGetPrimaryKeys:
		return handleRead(
			s, txnMeta, payload,
			func(
				req txnmemengine.GetPrimaryKeysReq,
				resp *txnmemengine.GetPrimaryKeysResp,
			) (
				err error,
			) {
				//TODO
				return
			},
		)

	case txnmemengine.OpGetTableDefs:
		return handleRead(
			s, txnMeta, payload,
			func(
				req txnmemengine.GetTableDefsReq,
				resp *txnmemengine.GetTableDefsResp,
			) (
				err error,
			) {
				//TODO
				return
			},
		)

	case txnmemengine.OpNewTableIter:
		return handleRead(
			s, txnMeta, payload,
			func(
				req txnmemengine.NewTableIterReq,
				resp *txnmemengine.NewTableIterResp,
			) (
				err error,
			) {
				//TODO
				return
			},
		)

	case txnmemengine.OpRead:
		return handleRead(
			s, txnMeta, payload,
			func(
				req txnmemengine.ReadReq,
				resp *txnmemengine.ReadResp,
			) (
				err error,
			) {
				//TODO
				return
			},
		)

	case txnmemengine.OpCloseTableIter:
		return handleRead(
			s, txnMeta, payload,
			func(
				req txnmemengine.CloseTableIterReq,
				resp *txnmemengine.CloseTableIterResp,
			) (
				err error,
			) {
				//TODO
				return
			},
		)

	}

	return
}

func handleRead[Req any, Resp any](
	s *Storage,
	txnMeta txn.TxnMeta,
	payload []byte,
	fn func(
		req Req,
		resp *Resp,
	) (
		err error,
	),
) (
	res storage.ReadResult,
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
	res = &readResult{
		payload: buf.Bytes(),
	}

	return res, nil
}
