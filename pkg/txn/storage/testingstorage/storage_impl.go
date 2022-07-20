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
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/storage"
)

var _ storage.TxnStorage = new(Storage)

func (s *Storage) StartRecovery(ch chan txn.TxnMeta) {
	//TODO
}

func (s *Storage) Read(txnMeta txn.TxnMeta, op uint32, payload []byte) (storage.ReadResult, error) {
	//TODO
	return nil, nil
}

func (s *Storage) Write(txnMeta txn.TxnMeta, op uint32, payload []byte) ([]byte, error) {
	//TODO
	return nil, nil
}

func (s *Storage) Prepare(txnMeta txn.TxnMeta) error {
	//TODO
	return nil
}

func (s *Storage) Committing(txnMeta txn.TxnMeta) error {
	//TODO
	return nil
}

func (s *Storage) Commit(txnMeta txn.TxnMeta) error {
	//TODO
	return nil
}

func (s *Storage) Rollback(txnMeta txn.TxnMeta) error {
	//TODO
	return nil
}
