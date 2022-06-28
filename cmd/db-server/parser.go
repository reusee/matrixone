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
	"strconv"
)

type Parser func(
	args []string,
) (
	nextArgs []string,
	cont Parser,
	err error,
)

func (p Parser) MatchStr(str string, cont Parser) Parser {
	return func(args []string) ([]string, Parser, error) {
		if len(args) == 0 {
			return nil, nil, fmt.Errorf("expecting %s, got nothing", str)
		}
		if args[0] != str {
			return nil, nil, fmt.Errorf("expecting %s, got %s", str, args[0])
		}
		return args[1:], cont, nil
	}
}

func (p Parser) Fallback(parser Parser, fallback Parser, fallbackArgs []string) Parser {
	return func(args []string) ([]string, Parser, error) {
		if parser == nil {
			return fallbackArgs, fallback, nil
		}
		nextArgs, nextParser, err := parser(args)
		if err != nil {
			return fallbackArgs, fallback, nil
		}
		if nextParser == nil {
			return nil, nil, nil
		}
		return nextArgs, p.Fallback(nextParser, fallback, fallbackArgs), nil
	}
}

func (p Parser) Alt(parsers ...Parser) Parser {
	if len(parsers) == 0 {
		return nil
	}
	return func(args []string) ([]string, Parser, error) {
		return args, p.Fallback(parsers[0], p.Alt(parsers[1:]...), args), nil
	}
}

func (p Parser) Repeat(repeating Parser, n int, cont Parser) Parser {
	if n == 0 || repeating == nil {
		return cont
	}
	parser := repeating
	var ret Parser
	ret = func(args []string) ([]string, Parser, error) {
		var err error
		args, parser, err = parser(args)
		if err != nil {
			return nil, nil, err
		}
		if parser == nil {
			if len(args) > 0 {
				return args, p.Repeat(repeating, n-1, cont), nil
			}
			return nil, cont, nil
		}
		return args, ret, nil
	}
	return ret
}

func (p Parser) End(fn func()) Parser {
	return func(args []string) ([]string, Parser, error) {
		fn()
		return args, nil, nil
	}
}

func (p Parser) Tap(fn func(string) error, cont Parser) Parser {
	return func(args []string) ([]string, Parser, error) {
		if len(args) == 0 {
			return nil, nil, fmt.Errorf("expecting string, got nothing")
		}
		if err := fn(args[0]); err != nil {
			return nil, nil, err
		}
		return args[1:], cont, nil
	}
}

func (p Parser) String(ptr *string, cont Parser) Parser {
	return p.Tap(func(str string) error {
		*ptr = str
		return nil
	}, cont)
}

func (p Parser) Uint64(ptr *uint64, cont Parser) Parser {
	return p.Tap(func(str string) error {
		num, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		*ptr = num
		return nil
	}, cont)
}

func (p Parser) Run(args []string) error {
	for {
		var err error
		args, p, err = p(args)
		if err != nil {
			return err
		}
		if p == nil {
			return nil
		}
	}
}
