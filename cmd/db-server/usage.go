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
	"fmt"
	"os"
	"sort"
)

type Usages []string

func (_ Def) HelpUsage() Usages {
	return Usages{
		`-h: this message`,
		`-help: this message`,
		`--help: this message`,
	}
}

func (_ Usages) IsReducer() {}

func (_ Def) Usages(
	usages Usages,
) (
	parsers ArgumentParsers,
) {

	sort.Strings(usages)

	var p Parser

	showUsages := p.End(func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		for _, line := range usages {
			fmt.Print(line)
			fmt.Print("\n")
		}
	})

	parsers = append(parsers, p.Seq(
		p.MatchStr("-h", nil),
		showUsages,
	))
	parsers = append(parsers, p.Seq(
		p.MatchStr("-help", nil),
		showUsages,
	))
	parsers = append(parsers, p.Seq(
		p.MatchStr("--help", nil),
		showUsages,
	))

	return
}
