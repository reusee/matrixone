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

import "testing"

func TestHandleArguments(t *testing.T) {
	var res []string

	NewScope().Fork(
		func() ArgumentParsers {
			var p Parser
			return ArgumentParsers{
				p.MatchStr("foo")(
					p.End(func() {
						res = append(res, "foo")
					})),
				p.MatchStr("bar")(
					p.End(func() {
						res = append(res, "bar")
					})),
			}
		},

		func() Arguments {
			return Arguments{
				"foo", "bar",
			}
		},
	).Call(func(
		handle HandleArguments,
	) {
		handle()
	})

	if len(res) != 2 {
		t.Fatalf("got %+v", res)
	}
	if res[0] != "foo" {
		t.Fatal()
	}
	if res[1] != "bar" {
		t.Fatal()
	}

}
