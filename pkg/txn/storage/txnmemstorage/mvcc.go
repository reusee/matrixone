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
	MinTxID         string
	MinPhysicalTime int64
	MinLogicalTime  int64
	MaxTxID         string
	MaxPhysicalTime int64
	MaxLogicalTime  int64
	Value           T
}

func (m *MVCC[T]) Read(
	txID string,
	physicalTime int64,
	logicalTime int64,
	txIsActive func(string) bool,
) *T {

	for i := len(m.Values) - 1; i >= 0; i-- {
		if m.Values[i].Visible(txID, physicalTime, logicalTime, txIsActive) {
			v := m.Values[i].Value
			return &v
		}
	}

	return nil
}

//TODO insert
//TODO delete
//TODO update

func (m *MVCCValue[T]) Visible(
	txID string,
	physicalTime int64,
	logicalTime int64,
	txIsActive func(string) bool,
) bool {

	// future value
	if physicalTime < m.MinPhysicalTime {
		return false
	}

	// future value
	if logicalTime < m.MinLogicalTime {
		return false
	}

	// uncommitted value by other tx
	if m.MinTxID != txID && txIsActive(m.MinTxID) {
		return false
	}

	// deleted by committed tx
	if m.MaxTxID != "" &&
		m.MaxTxID != txID &&
		(physicalTime > m.MaxPhysicalTime || (physicalTime == m.MaxPhysicalTime && logicalTime > m.MaxLogicalTime)) &&
		!txIsActive(m.MaxTxID) {
		return false
	}

	//TODO deleted by current tx

	return true
}
