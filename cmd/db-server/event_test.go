// Copyright 2021 Matrix Origin
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

package main

import (
	"testing"
)

func TestEvent(t *testing.T) {
	var res int

	scope := NewScope(func() int {
		return 1
	})

	scope.Call(func(
		on On,
	) {
		on("foo", func(
			i int,
		) {
			res = i
		})
	})

	scope = scope.Fork(func() int {
		return 2
	})

	scope.Call(func(
		emit Emit,
	) {
		emit(scope, "foo")
	})

	if res != 2 {
		t.Fatal()
	}

}
