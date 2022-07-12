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
	"reflect"
	"strings"
	"sync"
)

type Manager struct {
	parsers             []Parser
	positionalArguments []string
	usages              []Usage

	hooks struct {
		sync.Mutex
		Map map[string][]func()
	}
}

func NewManager() *Manager {
	m := new(Manager)
	m.hooks.Map = make(map[string][]func())

	// load specs
	v := reflect.ValueOf(m)
	for i := 0; i < v.NumMethod(); i++ {
		retValues := v.Method(i).Call([]reflect.Value{})
		for _, value := range retValues {
			switch ret := value.Interface().(type) {
			case Parser:
				m.parsers = append(m.parsers, ret)
			case []Parser:
				m.parsers = append(m.parsers, ret...)
			case Usage:
				m.usages = append(m.usages, ret)
			case []Usage:
				m.usages = append(m.usages, ret...)
			}
		}
	}

	return m
}

func (m *Manager) handleArguments(arguments []string) {

	//TODO fix this

	var p Parser
	resetParser := func() {
		p = p.AltElse(m.parsers, func(args []string) error {
			if len(args) > 1 {
				fmt.Printf("unknown arguments: %+v\n", args)
			} else if len(args) == 0 {
				m.printUsage()
				os.Exit(-1)
			}
			arg := args[0]
			if strings.HasPrefix(arg, "-") {
				// dash argument
				fmt.Printf("unknown argument: %s\n", arg)
				m.printUsage()
				os.Exit(-1)
			} else {
				// positional argument
				m.positionalArguments = append(m.positionalArguments, arg)
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
