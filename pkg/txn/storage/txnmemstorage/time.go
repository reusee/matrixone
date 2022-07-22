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

type Time [2]int64

func (t Time) Equal(t2 Time) bool {
	for i, n := range t {
		if n != t2[i] {
			return false
		}
	}
	return true
}

func (t Time) Before(t2 Time) bool {
	for i, n := range t {
		if n < t2[i] {
			return true
		}
		if n > t2[i] {
			return false
		}
	}
	return false
}

func (t Time) After(t2 Time) bool {
	for i, n := range t {
		if n > t2[i] {
			return true
		}
		if n < t2[i] {
			return false
		}
	}
	return false
}

var zeroTime Time

func (t Time) IsZero() bool {
	return t == zeroTime
}
