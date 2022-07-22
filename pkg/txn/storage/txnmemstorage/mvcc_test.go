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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMVCCRead(t *testing.T) {

	var txns []*Transaction
	for tn := 1; tn <= 5; tn++ {
		for _, state := range []TransactionState{Active, Committed, Aborted} {
			beginTime := Time{int64(tn), 0}
			var commitTime *Time
			if state != Active {
				commitTime = &Time{int64(tn), 5}
			}
			txns = append(txns, &Transaction{
				ID:         fmt.Sprintf("A time %d state %d", tn, state),
				BeginTime:  beginTime,
				CommitTime: commitTime,
				State:      state,
			})
			txns = append(txns, &Transaction{
				ID:         fmt.Sprintf("B time %d state %d", tn, state),
				BeginTime:  beginTime,
				CommitTime: commitTime,
				State:      state,
			})
		}
	}

	var values []*MVCC[int]
	for _, tx := range txns {
		values = append(values, &MVCC[int]{
			Values: []MVCCValue[int]{
				{
					CreateTx: tx,
					Value:    42,
				},
			},
		})
	}

	for _, tx := range txns {
		if tx.State != Active {
			// skip non-active tx
			continue
		}

		for _, value := range values {
			res := value.Read(tx)
			tx2 := value.Values[0].CreateTx

			if tx2 == tx {
				// read current tx created
				assert.NotNil(t, res)
				assert.Equal(t, 42, *res)
			}

			if tx2.State == Aborted {
				// read aborted
				assert.Nil(t, res)
			}

			if tx.BeginTime.Before(tx2.BeginTime) {
				// read future
				assert.Nil(t, res)
			}

			if tx2.CommitTime != nil && tx.BeginTime.Before(*tx2.CommitTime) {
				// read concurrent committed
				assert.Nil(t, res)
			}

			if tx2 != tx && tx2.State == Active {
				// read uncommitted
				assert.Nil(t, res)
			}

			if tx2.State == Committed && tx.BeginTime.After(*tx2.CommitTime) {
				// read committed
				assert.NotNil(t, res)
				assert.Equal(t, 42, *res)
			}

		}
	}

}
