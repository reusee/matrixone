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

package memstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type I int

func (i I) Less(i2 I) bool {
	return i < i2
}

func TestTable(t *testing.T) {
	table := NewTable[I, int]()

	tx := NewTransaction("1", Timestamp{})

	// insert
	err := table.Insert(tx, tx.CurrentTime, I(1), 1)
	assert.Nil(t, err)

	// get
	n, err := table.Get(tx, tx.CurrentTime, I(1))
	assert.Nil(t, err)
	assert.Equal(t, 1, n)

	// update
	err = table.Update(tx, tx.CurrentTime, I(1), 42)
	assert.Nil(t, err)

	n, err = table.Get(tx, tx.CurrentTime, I(1))
	assert.Nil(t, err)
	assert.Equal(t, 42, n)

	// delete
	err = table.Delete(tx, tx.CurrentTime, I(1))
	assert.Nil(t, err)

}
