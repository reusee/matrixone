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
	CreateTx *Transaction
	DeleteTx *Transaction
	Value    T
}

func (m *MVCC[T]) Read(tx *Transaction) *T {
	if tx.State != Active {
		panic("should not read")
	}

	for i := len(m.Values) - 1; i >= 0; i-- {
		if m.Values[i].Visible(tx) {
			if m.Values[i].DeleteTx != nil {
				// deleted
				return nil
			}
			v := m.Values[i].Value
			return &v
		}
	}

	return nil
}

//TODO insert
//TODO delete
//TODO update

func (m *MVCCValue[T]) Visible(tx *Transaction) bool {

	// read current tx created
	if tx == m.CreateTx {
		return true
	}

	// read committed not deleted
	if m.CreateTx.State == Committed &&
		tx.BeginTime.After(*m.CreateTx.CommitTime) &&
		m.DeleteTx == nil {
		return true
	}

	// read committed before deleted
	if m.CreateTx.State == Committed &&
		tx.BeginTime.After(*m.CreateTx.CommitTime) &&
		m.DeleteTx != nil &&
		tx.BeginTime.Before(m.DeleteTx.BeginTime) {
		return true
	}

	return false
}
