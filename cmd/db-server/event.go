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

import "sync"

const (
	evInit = "init"
	evExit = "exit"
)

type On func(ev string, fn func())

type Emit func(ev string)

func (_ Def) Event() (
	on On,
	emit Emit,
) {

	var l sync.Mutex
	events := make(map[string][]func())

	on = func(ev string, fn func()) {
		l.Lock()
		defer l.Unlock()
		events[ev] = append(events[ev], fn)
	}

	emit = func(ev string) {
		l.Lock()
		evs := events[ev]
		l.Unlock()
		for _, fn := range evs {
			fn()
		}
	}

	return
}
