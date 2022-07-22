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

type MVCC[T any] struct {
	Values []MVCCValue[T]
}

type MVCCValue[T any] struct {
	BornTx   *Transaction
	BornTime Timestamp
	LockTx   *Transaction
	LockTime Timestamp
	Value    T
}

// Read reads the visible value from Values
// readTime's logical time should be monotonically increasing in one transaction to reflect commands order
func (m *MVCC[T]) Read(tx *Transaction, readTime Timestamp) *T {
	if tx.State != Active {
		panic("should not read")
	}

	for i := len(m.Values) - 1; i >= 0; i-- {
		if m.Values[i].Visible(tx.ID, readTime) {
			v := m.Values[i].Value
			return &v
		}
	}

	return nil
}

//TODO insert
//TODO delete
//TODO update

func (m *MVCCValue[T]) Visible(txID string, readTime Timestamp) bool {

	// the following algorithm is from https://momjian.us/main/writings/pgsql/mvcc.pdf
	// "[Mike Olson] says 17 march 1993: the tests in this routine are correct; if you think they’re not, you’re wrongand you should think about it again. i know, it happened to me."

	// inserted by current tx
	if m.BornTx.ID == txID {
		// inserted before the read time
		if m.BornTime.Less(readTime) {
			// not been deleted
			if m.LockTx == nil {
				return true
			}
			// deleted by current tx after the read time
			if m.LockTx.ID == txID && m.LockTime.Greater(readTime) {
				return true
			}
		}
	}

	// inserted by a committed tx
	if m.BornTx.State == Committed {
		// not been deleted
		if m.LockTx == nil {
			return true
		}
		// being deleted by current tx after the read time
		if m.LockTx.ID == txID && m.LockTime.Greater(readTime) {
			return true
		}
		// deleted by another tx but not committed
		if m.LockTx.ID != txID && m.LockTx.State != Committed {
			return true
		}
	}

	return false
}
