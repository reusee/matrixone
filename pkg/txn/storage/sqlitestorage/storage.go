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
	"database/sql"

	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/storage"
)

type Storage struct {
	db *sql.DB
}

var _ storage.TxnStorage = new(Storage)

func (*Storage) Commit(txnMeta txn.TxnMeta) error {
	//TODO
	panic("unimplemented")
}

func (*Storage) Committing(txnMeta txn.TxnMeta) error {
	//TODO
	panic("unimplemented")
}

func (*Storage) Prepare(txnMeta txn.TxnMeta) error {
	//TODO
	panic("unimplemented")
}

func (*Storage) Read(txnMeta txn.TxnMeta, op uint32, payload []byte) (storage.ReadResult, error) {
	//TODO
	panic("unimplemented")
}

func (*Storage) Rollback(txnMeta txn.TxnMeta) error {
	//TODO
	panic("unimplemented")
}

func (*Storage) StartRecovery(chan txn.TxnMeta) {
	//TODO
	panic("unimplemented")
}

func (*Storage) Write(txnMeta txn.TxnMeta, op uint32, payload []byte) ([]byte, error) {
	//TODO
	panic("unimplemented")
}
