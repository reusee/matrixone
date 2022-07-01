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
	"strings"
)

type Usages [][2]string

func (_ Def) HelpUsage() Usages {
	return Usages{
		{"help | -h | -help | --help", "this message"},
	}
}

func (_ Usages) IsReducer() {}

type PrintUsages func()

func (_ Def) Usages(
	usages Usages,
) (
	parsers ArgumentParsers,
	printUsages PrintUsages,
) {

	sort.Slice(usages, func(i, j int) bool {
		keyA := usages[i][0]
		keyB := usages[j][0]
		var categoryA, categoryB int
		if strings.HasPrefix(keyA, "-") {
			categoryA = 1
		}
		if strings.HasPrefix(keyB, "-") {
			categoryB = 1
		}
		if categoryA != categoryB {
			return categoryA < categoryB
		}
		return keyA < keyB
	})

	printUsages = func() {
		fmt.Printf("usage: %s config_file_path\n", os.Args[0])
		maxLen := 0
		for _, pair := range usages {
			if l := len(pair[0]); l > maxLen {
				maxLen = l
			}
		}
		format := fmt.Sprintf("  %%%ds: %%s\n", maxLen)
		for _, pair := range usages {
			fmt.Printf(format, pair[0], pair[1])
		}
	}

	var p Parser

	parsers = append(parsers, p.MatchAnyStr("help", "-h", "-help", "--help")(
		p.End(func() {
			printUsages()
			os.Exit(0)
		})))

	return
}
