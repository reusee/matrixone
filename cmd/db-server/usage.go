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

type Usage struct {
	Header string
	Desc   string
}

func (m *Manager) printUsage() {

	sort.Slice(m.usages, func(i, j int) bool {
		keyA := m.usages[i].Header
		keyB := m.usages[j].Header
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

	fmt.Printf("usage: %s config_file_path\n\n", os.Args[0])
	maxLen := 0
	for _, usage := range m.usages {
		if l := len(usage.Header); l > maxLen {
			maxLen = l
		}
	}
	format := fmt.Sprintf("  %%-%ds    %%s\n", maxLen)
	for _, usage := range m.usages {
		fmt.Printf(format, usage.Header, usage.Desc)
	}
	fmt.Printf("\n")

}

func (m *Manager) Usage() (
	parser Parser,
	usage Usage,
) {

	var p Parser
	parser = p.MatchAnyStr("help", "-h", "-help", "--help")(
		p.End(func() {
			m.printUsage()
			os.Exit(0)
		}))

	usage.Header = "help | -h | -help | --help"
	usage.Desc = "this message"

	return
}
