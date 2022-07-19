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
	"github.com/matrixorigin/matrixone/pkg/txn/storage"
)

func decodePayload[T any](payload []byte) (value T, err error) {
	if err = gob.NewDecoder(bytes.NewReader(payload)).Decode(&value); err != nil {
		return
	}
	return
}

func encodeToReadResult(value any) (*readResult, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(value); err != nil {
		return nil, err
	}
	return &readResult{
		payload: buf.Bytes(),
	}, nil
}

type readResult struct {
	payload []byte
}

var _ storage.ReadResult = new(readResult)

func (r *readResult) Read() ([]byte, error) {
	return r.payload, nil
}

func (*readResult) Release() {
}

func (*readResult) WaitTxns() [][]byte {
	//TODO
	return nil
}

func handleRead[Req any, Resp any](
	s *Storage,
	txnMeta txn.TxnMeta,
	payload []byte,
	fn func(
		req Req,
	) (
		resp Resp,
		err error,
	),
) (
	res storage.ReadResult,
	err error,
) {

	req, err := decodePayload[Req](payload)
	if err != nil {
		return nil, err
	}

	resp, err := fn(req)
	if err != nil {
		return nil, err
	}

	res, err = encodeToReadResult(resp)
	if err != nil {
		return nil, err
	}

	return res, nil
}
