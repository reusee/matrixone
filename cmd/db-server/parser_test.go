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
		var s string
		var p Parser
		if err := p.Alt(
			p.MatchStr("1", p.End(func() {
				s = "1"
			})),
			p.MatchStr("foo", p.End(func() {
				s = "foo"
			})),
		).Run([]string{"foo"}); err != nil {
			t.Fatal(err)
		}
		if s != "foo" {
			t.Fatal()
		}
	})

	t.Run("Seq", func(t *testing.T) {
		var p Parser
		var a, b, c string
		if err := p.Seq(
			p.String(&a, nil),
			p.String(&b, nil),
			p.String(&c, nil),
		).Run([]string{"a", "b", "c"}); err != nil {
			t.Fatal(err)
		}
		if a != "a" {
			t.Fatal()
		}
		if b != "b" {
			t.Fatal()
		}
		if c != "c" {
			t.Fatal()
		}
	})

}
