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
	CreateTx   *Transaction
	CreateTime Time
	DeleteTx   *Transaction
	DeleteTime Time
	Value      T
}

// Read reads the visible value from Values
// readTime's logical time should be monotonically increasing in one transaction to reflect commands order
func (m *MVCC[T]) Read(tx *Transaction, readTime Time) *T {
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

func (m *MVCCValue[T]) Visible(txID string, readTime Time) bool {

	// the following algorithm is from https://momjian.us/main/writings/pgsql/mvcc.pdf
	// "[Mike Olson] says 17 march 1993: the tests in this routine are correct; if you think they’re not, you’re wrongand you should think about it again. i know, it happened to me."

	//TODO make it readable

	if m.CreateTx.ID == txID /* inserted by current tx */ &&
		m.CreateTime.Before(readTime) /* before the read time */ &&
		(m.DeleteTx == nil /* not been deleted, or */ ||
			(m.DeleteTx.ID == txID /* deleted by current tx */ &&
				m.DeleteTime.After(readTime) /* but after the read time */)) {
		return true
	}

	if m.CreateTx.State == Committed /* inserted by a committed tx */ &&
		(m.DeleteTx == nil /* not been deleted */ ||
			(m.DeleteTx.ID == txID /* being deleted by current tx */ &&
				m.DeleteTime.After(readTime) /* but after the read time */) ||
			(m.DeleteTx.ID != txID /* deleted by another tx */ &&
				m.DeleteTx.State != Committed /* but not committed */)) {
		return true
	}

	return false
}
