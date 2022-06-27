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
)

func (_ Def) Version() (
	parsers ArgumentParsers,
) {

	var p Parser
	parsers = ArgumentParsers{
		p.MatchStr("--version", p.End(func() {
			// if the argument passed in is "--version", return version info and exit
			fmt.Println("MatrixOne build info:")
			fmt.Printf("  The golang version used to build this binary: %s\n", GoVersion)
			fmt.Printf("  Git branch name: %s\n", BranchName)
			fmt.Printf("  Last git commit ID: %s\n", LastCommitId)
			fmt.Printf("  Buildtime: %s\n", BuildTime)
			fmt.Printf("  Current Matrixone version: %s\n", MoVersion)
			os.Exit(0)
		})),
	}

	return
}
