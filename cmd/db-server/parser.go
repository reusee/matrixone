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

func (p Parser) MatchStr(str string) Parser {
	return func(i *string) (Parser, error) {
		if i == nil {
			return nil, fmt.Errorf("expecting %s, got nothing", str)
		}
		if *i != str {
			return nil, fmt.Errorf("expecting %s, got %s", str, *i)
		}
		return nil, nil
	}
}

func (p Parser) MatchAnyStr(strs []string) Parser {
	return func(i *string) (Parser, error) {
		if i == nil {
			return nil, fmt.Errorf("expecting any of %+v, got nothing", strs)
		}
		for _, str := range strs {
			if *i == str {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("expecting any of %+v, got %s", strs, *i)
	}
}

func (p Parser) Alt(ps ...Parser) Parser {
	return p.AltElse(ps, func(_ *string) (Parser, error) {
		return nil, fmt.Errorf("no match")
	})
}

func (p Parser) AltElse(ps []Parser, elseParser Parser) Parser {
	parsers := make([]Parser, len(ps))
	copy(parsers, ps)
	var inputs []string
	var ret Parser
	ret = func(i *string) (Parser, error) {
		if i != nil && elseParser != nil {
			inputs = append(inputs, *i)
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
		if len(parsers) == 0 {
			var err error
			for _, input := range inputs {
				if elseParser == nil {
					return nil, nil
				}
				elseParser, err = elseParser(&input)
				if err != nil {
					return nil, err
				}
			}
			return elseParser, nil
		}
		if len(parsers) == 1 {
			next := parsers[0]
			if next != nil {
				return next(i)
			}
			return nil, nil
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

func (p Parser) Tap(fn func(string) error) Parser {
	return func(i *string) (Parser, error) {
		if i == nil {
			return nil, fmt.Errorf("expecting input")
		}
		if err := fn(*i); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func (p Parser) String(ptr *string) Parser {
	return p.Tap(func(str string) error {
		*ptr = str
		return nil
	})
}

func (p Parser) Uint64(ptr *uint64) Parser {
	return p.Tap(func(str string) error {
		num, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		*ptr = num
		return nil
	})
}

func (p Parser) Run(args []string) error {
	for {
		var input *string
		if len(args) > 0 {
			input = &args[0]
			args = args[1:]
		}
		if p == nil {
			break
		}
		var err error
		p, err = p(input)
		if err != nil {
			return err
		}
	}
	return nil
}
