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
	"sync"

	iradix "github.com/hashicorp/go-immutable-radix"
)

type Transaction struct {
	ID        string
	Snapshot  *iradix.Tree
	Log       []ChangeLogItem
	ChangeSet ChangeSet
}

type Transactions struct {
	sync.Mutex
	Map map[string]*Transaction
}

func (t *Transaction) Write() {
	//TODO update Snapshot
	//TODO update Log
	//TODO update global changeset
	//TODO do we need local changeset?
}

func (t *Transaction) Commit() {
	//TODO apply Log to main
	//TODO update main change history
}
