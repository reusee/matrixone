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
	"errors"
	"fmt"
	"strconv"
)

type Parser func(
	input *string,
) (
	next Parser,
	err error,
)

func (p Parser) MatchStr(str string) func(Parser) Parser {
	return func(cont Parser) Parser {
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
}

func (p Parser) MatchAnyStr(strs ...string) func(Parser) Parser {
	return func(cont Parser) Parser {
		return func(i *string) (Parser, error) {
			if i == nil {
				return nil, fmt.Errorf("expecting any of %+v, got nothing", strs)
			}
			for _, str := range strs {
				if *i == str {
					return cont, nil
				}
			}
			return nil, fmt.Errorf("expecting any of %+v, got %s", strs, *i)
		}
	}
}

func (p Parser) Alt(ps ...Parser) Parser {
	return p.AltElse(ps, func(_ []string) error {
		return fmt.Errorf("no match")
	})
}

func (p Parser) AltElse(ps []Parser, elseFunc func([]string) error) Parser {
	parsers := make([]Parser, len(ps))
	copy(parsers, ps)
	var inputs []string
	var ret Parser
	ret = func(i *string) (Parser, error) {
		if i != nil && elseFunc != nil {
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
			// no match
			if elseFunc != nil {
				if err := elseFunc(inputs); err != nil {
					return nil, err
				}
			}
			return nil, nil
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

func (p Parser) End(fn func()) Parser {
	return func(i *string) (Parser, error) {
		fn()
		if i != nil {
			return nil, InputAgain(*i)
		}
		return nil, nil
	}
}

func (p Parser) Tap(fn func(string) error) func(Parser) Parser {
	return func(cont Parser) Parser {
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
}

func (p Parser) String(ptr *string) func(Parser) Parser {
	return func(cont Parser) Parser {
		return p.Tap(func(str string) error {
			*ptr = str
			return nil
		})(cont)
	}
}

func (p Parser) Uint64(ptr *uint64) func(Parser) Parser {
	return func(cont Parser) Parser {
		return p.Tap(func(str string) error {
			num, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				return err
			}
			*ptr = num
			return nil
		})(cont)
	}
}

func (p Parser) Seq(parsers ...Parser) func(Parser) Parser {
	return func(cont Parser) Parser {
		var ret Parser
		ret = func(i *string) (Parser, error) {
			if len(parsers) == 0 {
				return cont.Consume(i)
			}
			for parsers[0] == nil {
				parsers = parsers[1:]
				if len(parsers) == 0 {
					return cont.Consume(i)
				}
			}
			var err error
			parsers[0], err = parsers[0](i)
			if err != nil {
				return nil, err
			}
			return ret, nil
		}
		return ret
	}
}

func (p Parser) First(parsers ...Parser) func(Parser) Parser {
	// strip nil parsers
	for i := 0; i < len(parsers); {
		if parsers[i] == nil {
			parsers[i] = parsers[len(parsers)-1]
			parsers = parsers[:len(parsers)-1]
		} else {
			i++
		}
	}
	return func(cont Parser) Parser {
		var ret Parser
		ret = func(input *string) (Parser, error) {
			for i := 0; i < len(parsers); {
				parser := parsers[i]
				var err error
				parser, err = parser(input)
				if err != nil {
					// skip
					parsers[i] = parsers[len(parsers)-1]
					parsers = parsers[:len(parsers)-1]
					continue
				}
				if parser == nil {
					// end
					return cont, nil
				}
				parsers[i] = parser
				i++
			}
			if len(parsers) == 0 {
				return nil, fmt.Errorf("no match")
			}
			return ret, nil
		}
		return ret
	}
}

func (p Parser) Longest(parsers ...Parser) Parser {
	// strip nil parsers
	for i := 0; i < len(parsers); {
		if parsers[i] == nil {
			parsers[i] = parsers[len(parsers)-1]
			parsers = parsers[:len(parsers)-1]
		} else {
			i++
		}
	}
	var ret Parser
	ret = func(input *string) (Parser, error) {
		for i := 0; i < len(parsers); {
			parser := parsers[i]
			var err error
			parser, err = parser(input)
			if err != nil || parser == nil {
				// skip
				parsers[i] = parsers[len(parsers)-1]
				parsers = parsers[:len(parsers)-1]
				continue
			}
			parsers[i] = parser
			i++
		}
		if len(parsers) == 0 {
			if input == nil {
				return nil, nil
			}
			return nil, fmt.Errorf("no match")
		}
		if len(parsers) == 1 {
			// the one
			return parsers[0], nil
		}
		return ret, nil
	}
	return ret
}

func (p Parser) OneOrMore(parser Parser) func(Parser) Parser {
	return func(cont Parser) Parser {
		var split Parser
		split = func(i *string) (Parser, error) {
			return p.Longest(
				p.Seq(parser)(cont),
				p.Seq(parser)(split),
			).Consume(i)
		}
		return split
	}
}

func (p Parser) Run(args []string) error {
	for {
		var input *string
		if len(args) > 0 {
			input = &args[0]
			args = args[1:]
		}
	l1:
		if p == nil {
			break
		}
		var err error
		p, err = p(input)
		if err != nil {
			var again InputAgain
			if errors.As(err, &again) {
				str := string(again)
				input = &str
				goto l1
			}
			return err
		}
	}
	return nil
}

func (p Parser) Consume(i *string) (Parser, error) {
	if p == nil {
		return nil, nil
	}
	return p(i)
}

type InputAgain string

func (i InputAgain) Error() string {
	return "input again: " + string(i)
}
