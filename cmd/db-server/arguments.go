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
	"strings"
)

type ArgumentParsers []Parser

func (_ ArgumentParsers) IsReducer() {}

func (_ Def) ArgumentParsers() ArgumentParsers {
	return nil
}

type Arguments []string

func (_ Def) Arguments() Arguments {
	return Arguments(os.Args[1:])
}

type PositionalArguments *[]string

func (_ Def) PositionalArguments() PositionalArguments {
	return &[]string{}
}

type HandleArguments func()

func (_ Def) HandleArguments(
	parsers ArgumentParsers,
	arguments Arguments,
	printUsages PrintUsages,
	posArgs PositionalArguments,
) (
	handle HandleArguments,
) {

	handle = func() {

		var p Parser
		resetParser := func() {
			p = p.AltElse(parsers, func(args []string) error {
				if len(args) > 1 {
					fmt.Printf("unknown arguments: %+v\n", args)
				}
				arg := args[0]
				if strings.HasPrefix(arg, "-") {
					// dash argument
					fmt.Printf("unknown argument: %s\n", arg)
					printUsages()
					os.Exit(-1)
				} else {
					// positional argument
					*posArgs = append(*posArgs, arg)
				}
				return nil
			})
		}
		resetParser()

		for {
			if len(arguments) == 0 {
				break
			}
			input := &arguments[0]
			arguments = arguments[1:]
			if p == nil {
				resetParser()
			}
			var err error
			p, err = p(input)
			if err != nil {
				panic(err)
			}
		}

		for p != nil {
			var err error
			p, err = p(nil)
			if err != nil {
				panic(err)
			}
		}

	}

	return
}
