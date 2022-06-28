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
	"os"
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

type HandleArguments func()

func (_ Def) HandleArguments(
	parsers ArgumentParsers,
	arguments Arguments,
	printUsages PrintUsages,
) HandleArguments {
	return func() {

		var loop Parser
		var p Parser
		loop = func(i *string) (Parser, error) {
			if p == nil {
				p = p.AltElse(parsers, p.End(func() {
					printUsages()
				}))
			}
			var err error
			p, err = p(i)
			if err != nil {
				return nil, err
			}
			if i == nil {
				return p, nil
			}
			return loop, nil
		}

		if err := loop.Run(arguments); err != nil {
			panic(err)
		}

	}
}
