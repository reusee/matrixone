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

	t.Run("Alt", func(t *testing.T) {
		var ok bool
		var p Parser
		if err := p.Alt(
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

	t.Run("Repeat", func(t *testing.T) {
		var n int
		var p Parser
		if err := p.Repeat(p.End(func() {
			n++
		}), 3, nil).Run([]string{"foo", "bar", "baz", "qux"}); err != nil {
			t.Fatal(err)
		}
		if n != 3 {
			t.Fatalf("got %d", n)
		}
	})

	t.Run("Repeat2", func(t *testing.T) {
		var p Parser
		var strs []string
		if err := p.Repeat(p.Tap(func(s string) error {
			strs = append(strs, s)
			return nil
		}, nil), -1, nil).Run([]string{"foo", "bar"}); err != nil {
			t.Fatal(err)
		}
		if len(strs) != 2 {
			t.Fatal()
		}
	})

}
