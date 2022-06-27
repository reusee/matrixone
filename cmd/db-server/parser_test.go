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

func TestParser(t *testing.T) {

	t.Run("MatchStr", func(t *testing.T) {
		var ok bool
		var p Parser
		if err := p.MatchStr("foo", p.End(func() {
			ok = true
		})).Run([]string{"foo"}); err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal()
		}
	})

	t.Run("MatchStr2", func(t *testing.T) {
		var ok bool
		var p Parser
		if err := p.MatchStr(
			"foo",
			p.MatchStr(
				"bar",
				p.End(func() {
					ok = true
				}),
			),
		).Run([]string{"foo", "bar"}); err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal()
		}
	})

	t.Run("Fallback", func(t *testing.T) {
		var ok bool
		var p Parser
		args := []string{"bar"}
		if err := p.Fallback(
			p.MatchStr("foo", nil),
			p.MatchStr("bar", p.End(func() {
				ok = true
			})),
			args,
		).Run(args); err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal()
		}
	})

	t.Run("First", func(t *testing.T) {
		var ok bool
		var p Parser
		if err := p.First(
			p.MatchStr("1", p.End(func() {
				ok = true
			})),
			p.MatchStr("foo", p.End(func() {
				ok = true
			})),
		).Run([]string{"foo"}); err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal()
		}
	})

}
