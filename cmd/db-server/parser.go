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
	input *string,
) (
	cont Parser,
	err error,
)

func (p Parser) MatchStr(str string, cont Parser) Parser {
	return func(i *string) (Parser, error) {
		if i == nil {
			return nil, fmt.Errorf("expecting %s, got nothing", str)
		}
		if *i != str {
			return nil, fmt.Errorf("expecting %s, got %s", str, *i)
		}
		return cont, nil
	}
}

func (p Parser) Alt(parsers ...Parser) Parser {
	if len(parsers) == 0 {
		return nil
	}
	return func(i *string) (Parser, error) {
		if i == nil && len(parsers) == 0 {
			return nil, nil
		}
		if len(parsers) == 0 {
			return nil, fmt.Errorf("no match")
		}
		if len(parsers) == 1 {
			return parsers[0], nil
		}
		for n := 0; n < len(parsers); {
			parser, err := parsers[n](i)
			if err != nil || parser == nil {
				parsers[n] = parsers[len(parsers)-1]
				parsers = parsers[:len(parsers)-1]
				continue
			}
			parsers[n] = parser
			n++
		}
		return p.Alt(parsers...), nil
	}
}

func (p Parser) Repeat(repeating Parser, n int, cont Parser) Parser {
	if n == 0 || repeating == nil {
		return cont
	}
	parser := repeating
	var ret Parser
	ret = func(i *string) (Parser, error) {
		if i == nil {
			return cont, nil
		}
		var err error
		parser, err = parser(i)
		if err != nil {
			return nil, err
		}
		if parser == nil {
			return p.Repeat(repeating, n-1, cont), nil
		}
		return ret, nil
	}
	return ret
}

func (p Parser) Seq(parsers ...Parser) Parser {
	if len(parsers) == 0 {
		return nil
	}
	parser := parsers[0]
	parsers = parsers[1:]
	var ret Parser
	ret = func(i *string) (Parser, error) {
		if parser == nil {
			next := p.Seq(parsers...)
			if next != nil {
				return next(i)
			}
			return nil, nil
		}
		var err error
		parser, err = parser(i)
		if err != nil {
			return nil, err
		}
		return ret, nil
	}
	return ret
}

func (p Parser) End(fn func()) Parser {
	return func(i *string) (Parser, error) {
		fn()
		return nil, nil
	}
}

func (p Parser) Tap(fn func(string) error, cont Parser) Parser {
	return func(i *string) (Parser, error) {
		if i == nil {
			return nil, fmt.Errorf("expecting input")
		}
		if err := fn(*i); err != nil {
			return nil, err
		}
		return cont, nil
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
	for _, input := range args {
		if p == nil {
			break
		}
		var err error
		p, err = p(&input)
		if err != nil {
			return err
		}
	}
	for p != nil {
		var err error
		p, err = p(nil)
		if err != nil {
			return err
		}
	}
	return nil
}
