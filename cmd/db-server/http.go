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
	"net/http"
	_ "net/http/pprof"
)

func (_ Def) HTTP(
	on On,
) (
	parsers ArgumentParsers,
	usages Usages,
) {

	var p Parser
	var addr string
	parsers = append(parsers, p.MatchStr("-http")(
		p.String(&addr)(
			p.End(func() {
				on(evInit, func() {
					go startHTTPServer(addr)
				})
			}))))
	usages = append(usages, [2]string{`-http address`, `start http server at specified address`})

	return
}

func startHTTPServer(addr string) {
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
