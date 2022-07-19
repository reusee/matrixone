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
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/txnmemengine"
)

func (*Storage) Write(txnMeta txn.TxnMeta, op uint32, payload []byte) (result []byte, err error) {

	switch op {

	case txnmemengine.OpCreateDatabase:
		//TODO

	case txnmemengine.OpDeleteDatabase:
		//TODO

	case txnmemengine.OpCreateRelation:
		//TODO

	case txnmemengine.OpDeleteRelation:
		//TODO

	case txnmemengine.OpAddTableDef:
		//TODO

	case txnmemengine.OpDelTableDef:
		//TODO

	case txnmemengine.OpDelete:
		//TODO

	case txnmemengine.OpTruncate:
		//TODO

	case txnmemengine.OpUpdate:
		//TODO

	case txnmemengine.OpWrite:
		//TODO

	}

	return
}
