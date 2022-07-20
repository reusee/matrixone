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

type Storage struct {
	main          *iradix.Tree
	changeHistory []ChangeHistoryItem
	changeSet     *GlobalChangeSet
	transactions  *Transactions
}

type Key any

type ChangeSet struct {
	Reads  map[Key]struct{}
	Writes map[Key]struct{}
}

func NewChangeSet() *ChangeSet {
	return &ChangeSet{
		Reads:  make(map[Key]struct{}),
		Writes: make(map[Key]struct{}),
	}
}

type ChangeHistoryItem struct {
	Version   int64
	Changeset ChangeSet
}

type ChangeLogItem struct {
	Key   any
	Value any
}

type GlobalChangeSet struct {
	sync.RWMutex
	Reads  map[Key]map[*Transaction]struct{}
	Writes map[Key]map[*Transaction]struct{}
}
